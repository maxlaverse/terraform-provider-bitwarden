//go:build offline

package provider

import (
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureClientsMuxHandoff(t *testing.T) {
	cfg := providerConfig{
		Server:               "http://127.0.0.1/",
		Email:                "mux-handoff@laverse.net",
		ClientID:             "client-id-mux",
		ClientSecret:         "client-secret-mux",
		MasterPassword:       "master-password-mux",
		VaultPath:            explicitVaultPath(t.TempDir()),
		ClientImplementation: schema_definition.ClientImplementationEmbedded,
	}

	offered, err := configureClientsOffer(t.Context(), versionTestSkippedLogin, cfg)
	require.NoError(t, err)
	require.NotNil(t, offered)

	taken, err := configureClientsTakeOrCreate(t.Context(), versionTestSkippedLogin, cfg)
	require.NoError(t, err)
	assert.Same(t, offered, taken, "SDKv2 Configure should reuse Framework clients for one mux RPC")

	// A later ConfigureProvider must not see the previous offer.
	again, err := configureClientsTakeOrCreate(t.Context(), versionTestSkippedLogin, cfg)
	require.NoError(t, err)
	assert.NotSame(t, offered, again, "consumed offer must not be reused across ConfigureProvider RPCs")
}

func TestConfigureClientsTakeWithoutOffer(t *testing.T) {
	cfg := providerConfig{
		Server:               "http://127.0.0.1/",
		Email:                "mux-solo@laverse.net",
		ClientID:             "client-id-solo",
		ClientSecret:         "client-secret-solo",
		MasterPassword:       "master-password-solo",
		VaultPath:            explicitVaultPath(t.TempDir()),
		ClientImplementation: schema_definition.ClientImplementationEmbedded,
	}

	first, err := configureClientsTakeOrCreate(t.Context(), versionTestSkippedLogin, cfg)
	require.NoError(t, err)
	second, err := configureClientsTakeOrCreate(t.Context(), versionTestSkippedLogin, cfg)
	require.NoError(t, err)
	assert.NotSame(t, first, second, "without an offer each take builds a fresh client")
}

func TestConfigureClientsOfferReplacesOrphan(t *testing.T) {
	cfg := providerConfig{
		Server:               "http://127.0.0.1/",
		Email:                "mux-orphan@laverse.net",
		ClientID:             "client-id-orphan",
		ClientSecret:         "client-secret-orphan",
		MasterPassword:       "master-password-orphan",
		VaultPath:            explicitVaultPath(t.TempDir()),
		ClientImplementation: schema_definition.ClientImplementationEmbedded,
	}

	orphan, err := configureClientsOffer(t.Context(), versionTestSkippedLogin, cfg)
	require.NoError(t, err)

	replacement, err := configureClientsOffer(t.Context(), versionTestSkippedLogin, cfg)
	require.NoError(t, err)
	assert.NotSame(t, orphan, replacement)

	taken, err := configureClientsTakeOrCreate(t.Context(), versionTestSkippedLogin, cfg)
	require.NoError(t, err)
	assert.Same(t, replacement, taken)
}
