//go:build offline

package provider

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwscli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEphemeralSecretsCLIBatch(t *testing.T) {
	for _, outcome := range []string{"success", "missing", "malformed", "failure", "not-found"} {
		t.Run(outcome, func(t *testing.T) {
			// Exercise the real CLI adapter and command runner using the existing
			// helper process, whose output also contains an unrelated secret.
			newCommand := command.New
			calls := 0
			command.New = func(binary string, args ...string) command.Command {
				calls++
				require.Equal(t, "bws", binary)
				require.Equal(t, []string{"secret", "list", "--output", "json"}, args)
				return newCommand(os.Args[0], append([]string{"-test.run=^TestEphemeralSecretCLIProcess$", "--"}, args...)...).AppendEnv([]string{
					"TEST_BACKEND=vaultwarden",
					"BITWARDEN_EPHEMERAL_TEST_OUTCOME=" + outcome,
				})
			}
			t.Cleanup(func() { command.New = newCommand })

			client := bwscli.NewSecretsManagerClient("http://127.0.0.1")
			require.NoError(t, client.LoginWithAccessToken(t.Context(), "test-access-token"))
			server := ephemeralSecretServer(t, &ProviderClients{SecretsManager: client})

			ids := []string{"secret-id", "SECRET-ID"}
			if outcome == "missing" {
				ids = append(ids, "missing-id")
			}

			var logs bytes.Buffer
			ctx := tflogtest.RootLogger(t.Context(), &logs)

			resp, err := server.OpenEphemeralResource(ctx, &tfprotov6.OpenEphemeralResourceRequest{
				TypeName: "bitwarden_secrets",
				Config:   ephemeralSecretsConfig(t, ids, nil),
			})
			require.NoError(t, err)

			assert.Equal(t, 1, calls)
			if outcome == "success" {
				require.Empty(t, resp.Diagnostics)
				assert.Equal(t, map[string]string{"secret-id": "private-value\"\n雪"}, ephemeralSecretsValues(t, resp))
			} else {
				require.Len(t, resp.Diagnostics, 1)
				assertEphemeralSecretsNoValues(t, resp)
				if outcome == "missing" || outcome == "not-found" {
					assert.Equal(t, models.ErrObjectNotFound.Error(), resp.Diagnostics[0].Detail)
				}
			}

			for _, marker := range []string{"private-value", "private-note", "unrelated-private", "private-stderr"} {
				assert.NotContains(t, logs.String(), marker)
				assert.NotContains(t, fmt.Sprint(resp.Diagnostics), marker)
			}
			assert.Contains(t, logs.String(), "Command finished")
			assert.Empty(t, resp.Private)

			tflog.Trace(ctx, "Logging scope control", map[string]any{"stdout": "ordinary-output-control"})
			assert.Contains(t, logs.String(), "ordinary-output-control")

			closed, err := server.CloseEphemeralResource(ctx, &tfprotov6.CloseEphemeralResourceRequest{
				TypeName: "bitwarden_secrets",
			})
			require.NoError(t, err)
			assert.Empty(t, closed.Diagnostics)

			resp, err = server.OpenEphemeralResource(ctx, &tfprotov6.OpenEphemeralResourceRequest{
				TypeName: "bitwarden_secrets",
				Config:   ephemeralSecretsConfig(t, []string{}, nil),
			})
			require.NoError(t, err)
			require.Empty(t, resp.Diagnostics)
			assert.Empty(t, ephemeralSecretsValues(t, resp))
			assert.Equal(t, 1, calls, "Close and empty batches must not invoke the CLI")
		})
	}
}
