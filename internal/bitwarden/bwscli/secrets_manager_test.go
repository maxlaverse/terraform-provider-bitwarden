//go:build offline

package bwscli

import (
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	test_command "github.com/maxlaverse/terraform-provider-bitwarden/internal/command/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSecretUnmarshalErrorOmitsOutput(t *testing.T) {
	removeMocks, _ := test_command.MockCommands(t, map[string]string{
		"secret get secret-id": `THIS_IS_A_SECRET`,
	})
	defer removeMocks(t)

	client := NewSecretsManagerClient("http://127.0.0.1")
	require.NoError(t, client.LoginWithAccessToken(t.Context(), "test-access-token"))

	_, err := client.GetSecret(t.Context(), models.Secret{ID: "secret-id"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "unable to parse result of 'secret get':")
	assert.NotContains(t, err.Error(), "THIS_IS_A_SECRET")
}
