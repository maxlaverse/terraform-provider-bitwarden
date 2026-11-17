//go:build integration

package provider

import (
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccEmbeddedClientDetectsServerType(t *testing.T) {
	ensureTestAccountsExist(t)

	client, ok := bwEmbeddedTestClient(t).(embedded.PasswordManagerClient)
	require.True(t, ok)

	expected := webapi.ServerTypeBitwarden
	if IsVaultwardenBackend() {
		expected = webapi.ServerTypeVaultwarden
	}

	assert.Equal(t, expected, client.ServerType())
	assert.NotEmpty(t, client.ServerVersion())
}
