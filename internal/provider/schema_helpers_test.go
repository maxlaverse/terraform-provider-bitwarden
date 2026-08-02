//go:build offline

package provider

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientsFromProviderData(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	clients, ok := clientsFromProviderData(nil, &diags)
	assert.False(t, ok)
	assert.Nil(t, clients)
	assert.False(t, diags.HasError())

	diags = nil
	clients, ok = clientsFromProviderData("wrong", &diags)
	assert.False(t, ok)
	assert.Nil(t, clients)
	assert.True(t, diags.HasError())

	expected := &ProviderClients{}
	diags = nil
	clients, ok = clientsFromProviderData(expected, &diags)
	assert.True(t, ok)
	assert.Same(t, expected, clients)
	assert.False(t, diags.HasError())
}

func TestRequirePasswordManagerMissing(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	client, ok := requirePasswordManager(&ProviderClients{}, &diags)
	assert.False(t, ok)
	assert.Nil(t, client)
	require.True(t, diags.HasError())
	assert.Equal(t, "Provider not configured for Password Manager", diags[0].Summary())
}

func TestRequireSecretsManagerMissing(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	client, ok := requireSecretsManager(&ProviderClients{}, &diags)
	assert.False(t, ok)
	assert.Nil(t, client)
	require.True(t, diags.HasError())
	assert.Equal(t, "Provider not configured for Secrets Manager", diags[0].Summary())
}

func TestAddErr(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	addErr(&diags, errors.New("object not found"))
	require.Len(t, diags, 1)
	assert.Equal(t, "object not found", diags[0].Summary())
	assert.Empty(t, diags[0].Detail())
}
