//go:build offline

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwscli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEphemeralSecretCLILogging(t *testing.T) {
	for _, selector := range []string{"id", "key"} {
		for _, outcome := range []string{"success", "malformed", "failure", "not-found"} {
			t.Run(selector+"/"+outcome, func(t *testing.T) {
				// Keep the real CLI adapter and command runner, replacing only
				// the external bws executable with this test's helper process.
				newCommand := command.New
				command.New = func(binary string, args ...string) command.Command {
					require.Equal(t, "bws", binary)
					return newCommand(os.Args[0], append([]string{"-test.run=^TestEphemeralSecretCLIProcess$", "--"}, args...)...).AppendEnv([]string{
						"TEST_BACKEND=vaultwarden",
						"BITWARDEN_EPHEMERAL_TEST_OUTCOME=" + outcome,
					})
				}
				t.Cleanup(func() { command.New = newCommand })
				client := bwscli.NewSecretsManagerClient("http://127.0.0.1")
				require.NoError(t, client.LoginWithAccessToken(t.Context(), "test-access-token"))
				server := ephemeralSecretServer(t, &ProviderClients{SecretsManager: client})

				var logs bytes.Buffer
				ctx := tflogtest.RootLogger(t.Context(), &logs)
				input := "secret-id"
				if selector == "key" {
					input = "secret-key"
				}
				resp, err := server.OpenEphemeralResource(ctx, &tfprotov6.OpenEphemeralResourceRequest{
					TypeName: "bitwarden_secret",
					Config:   ephemeralSecretConfig(t, map[string]any{selector: input}),
				})
				require.NoError(t, err)
				if outcome == "success" {
					require.Empty(t, resp.Diagnostics)
					assertEphemeralSecretValue(t, resp, "private-value\"\n雪")
				} else {
					require.Len(t, resp.Diagnostics, 1)
					if outcome == "not-found" {
						assert.Equal(t, models.ErrObjectNotFound.Error(), resp.Diagnostics[0].Detail)
					} else {
						assert.Contains(t, resp.Diagnostics[0].Detail, "Client error details are omitted")
					}
				}
				for _, marker := range []string{"private-value", "private-note", "unrelated-private", "private-stderr"} {
					assert.NotContains(t, logs.String(), marker)
					assert.NotContains(t, fmt.Sprint(resp.Diagnostics), marker)
				}
				assert.Contains(t, logs.String(), "Command finished", "command status should remain available")
				// The filter must be local to Open, not modify the caller's logger.
				tflog.Trace(ctx, "Logging scope control", map[string]any{"stdout": "ordinary-output-control"})
				assert.Contains(t, logs.String(), "ordinary-output-control")
			})
		}
	}
}

func TestEphemeralSecretCLIProcess(t *testing.T) {
	outcome := os.Getenv("BITWARDEN_EPHEMERAL_TEST_OUTCOME")
	if outcome == "" {
		return
	}

	secret := models.Secret{ID: "secret-id", Key: "secret-key", Value: "private-value\"\n雪", Note: "private-note"}
	var result any = secret
	for i, arg := range os.Args {
		if arg == "--" && i+2 < len(os.Args) && os.Args[i+2] == "list" {
			result = []models.Secret{secret, {ID: "other-id", Key: "other-key", Value: "unrelated-private"}}
			break
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		os.Exit(2)
	}
	// Even successful commands can put sensitive details on stderr.
	fmt.Fprintln(os.Stderr, "private-stderr")
	switch outcome {
	case "malformed":
		fmt.Fprintln(os.Stdout, "invalid trailing JSON: private-value")
	case "failure":
		os.Exit(1)
	case "not-found":
		fmt.Fprintln(os.Stderr, "Resource not found.")
		os.Exit(1)
	}
	os.Exit(0)
}

func TestEphemeralSecretEmbeddedDuplicateKey(t *testing.T) {
	backend := NewTestSecretsManager()
	router := mux.NewRouter()
	router.HandleFunc("/identity/connect/token", backend.handlerLogin).Methods("POST")
	router.HandleFunc("/api/organizations/{orgId}/projects", backend.handlerCreateProject).Methods("POST")
	router.HandleFunc("/api/organizations/{orgId}/secrets", backend.handlerCreateGetSecret).Methods("POST", "GET")
	router.HandleFunc("/api/secrets/{secretId}", backend.handlerGetSecret).Methods("GET")
	transport := httpmock.NewMockTransport()
	transport.RegisterNoResponder(func(req *http.Request) (*http.Response, error) {
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		return resp.Result(), nil
	})
	client := embedded.NewSecretsManagerClient("http://127.0.0.1", "test-device", "test",
		embedded.WithSecretsManagerHttpOptions(webapi.WithCustomClient(http.Client{Transport: transport})),
	)
	org, err := backend.ClientCreateNewOrganization()
	require.NoError(t, err)
	token, err := backend.ClientCreateAccessToken(org)
	require.NoError(t, err)
	require.NoError(t, client.LoginWithAccessToken(t.Context(), token))
	project, err := client.CreateProject(t.Context(), models.Project{Name: "test-project", OrganizationID: org})
	require.NoError(t, err)
	wanted := models.Secret{Key: "duplicate-key", Value: "private-value", Note: "private-note", ProjectID: project.ID, OrganizationID: org}
	secret, err := client.CreateSecret(t.Context(), wanted)
	require.NoError(t, err)
	server := ephemeralSecretServer(t, &ProviderClients{SecretsManager: client})

	// Both selectors must still read a unique secret through the real client.
	for _, attrs := range []map[string]any{{"id": secret.ID}, {"key": wanted.Key}} {
		resp, err := server.OpenEphemeralResource(t.Context(), &tfprotov6.OpenEphemeralResourceRequest{
			TypeName: "bitwarden_secret", Config: ephemeralSecretConfig(t, attrs),
		})
		require.NoError(t, err)
		require.Empty(t, resp.Diagnostics)
		assertEphemeralSecretValue(t, resp, wanted.Value)
	}

	_, err = client.CreateSecret(t.Context(), wanted)
	require.NoError(t, err)
	_, err = client.GetSecretByKey(t.Context(), wanted.Key)
	assert.ErrorIs(t, err, models.ErrTooManyObjectsFound)
	resp, err := server.OpenEphemeralResource(t.Context(), &tfprotov6.OpenEphemeralResourceRequest{
		TypeName: "bitwarden_secret", Config: ephemeralSecretConfig(t, map[string]any{"key": wanted.Key}),
	})
	require.NoError(t, err)
	require.Len(t, resp.Diagnostics, 1)
	assert.Equal(t, tfprotov6.DiagnosticSeverityError, resp.Diagnostics[0].Severity)
	assert.Equal(t, "too many objects found", resp.Diagnostics[0].Detail)
}

func assertEphemeralSecretValue(t *testing.T, resp *tfprotov6.OpenEphemeralResourceResponse, want string) {
	t.Helper()
	require.NotNil(t, resp.Result)
	result, err := resp.Result.Unmarshal(schema_definition.SecretEphemeralResourceSchema().Type().TerraformType(t.Context()))
	require.NoError(t, err)
	var attrs map[string]tftypes.Value
	require.NoError(t, result.As(&attrs))
	var value string
	require.NoError(t, attrs["value"].As(&value))
	assert.Equal(t, want, value)
}
