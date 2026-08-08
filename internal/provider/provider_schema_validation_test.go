//go:build offline

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
)

// TestProviderSchemaValidity ensures the muxed Framework+SDKv2 provider schemas
// align (terraform-plugin-mux rejects mismatched provider schemas). Migrated
// Framework types are asserted below; remaining SDKv2 types (including
// attachments) are covered by schema_contract_test.go.
func TestProviderSchemaValidity(t *testing.T) {
	factory, err := NewProviderServer(versionTestSkippedLogin)
	if err != nil {
		t.Fatalf("mux provider schemas must align: %s", err)
	}
	server := factory()

	resp, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("schema error: %s: %s", d.Summary, d.Detail)
		}
	}

	for _, name := range []string{"bitwarden_folder", "bitwarden_project", "bitwarden_secret"} {
		if _, ok := resp.ResourceSchemas[name]; !ok {
			t.Fatalf("expected Framework %s resource to be registered", name)
		}
	}
	for _, name := range []string{
		"bitwarden_folder",
		"bitwarden_project",
		"bitwarden_secret",
		"bitwarden_organization",
		"bitwarden_org_group",
		"bitwarden_org_member",
	} {
		if _, ok := resp.DataSourceSchemas[name]; !ok {
			t.Fatalf("expected Framework %s data source to be registered", name)
		}
	}
	if _, ok := resp.ResourceSchemas["bitwarden_attachment"]; !ok {
		t.Fatal("expected SDKv2 bitwarden_attachment resource to remain registered during mux migration")
	}
}

func TestProviderAuthUsingAPIKey(t *testing.T) {
	cfg := providerConfig{
		Server:         "http://127.0.0.1/",
		Email:          "test@laverse.net",
		ClientID:       "client-id-1234",
		ClientSecret:   "client-secret-5678",
		MasterPassword: "master-password-9",
	}
	assert.NoError(t, validateProviderConfig(cfg))

	clients, err := configureClients(t.Context(), versionTestSkippedLogin, cfg)
	assert.NoError(t, err)

	pm, err := clients.RequirePasswordManager()
	assert.NoError(t, err)
	assert.Implements(t, (*bwcli.PasswordManagerClient)(nil), pm)
}

func TestProviderAuthUsingAPIAndEmbedded(t *testing.T) {
	cfg := providerConfig{
		Server:               "http://127.0.0.1/",
		Email:                "test@laverse.net",
		ClientID:             "client-id-1234",
		ClientSecret:         "client-secret-5678",
		MasterPassword:       "master-password-9",
		ClientImplementation: schema_definition.ClientImplementationEmbedded,
	}
	assert.NoError(t, validateProviderConfig(cfg))

	clients, err := configureClients(t.Context(), versionTestSkippedLogin, cfg)
	assert.NoError(t, err)

	pm, err := clients.RequirePasswordManager()
	assert.NoError(t, err)
	assert.Implements(t, (*embedded.PasswordManagerClient)(nil), pm)
}

func TestProviderAuthUsingSessionKey(t *testing.T) {
	cfg := providerConfig{
		Server:     "http://127.0.0.1/",
		Email:      "test@laverse.net",
		SessionKey: "1234",
	}
	assert.NoError(t, validateProviderConfig(cfg))

	clients, err := configureClients(t.Context(), versionTestSkippedLogin, cfg)
	assert.NoError(t, err)

	pm, err := clients.RequirePasswordManager()
	assert.NoError(t, err)
	assert.Implements(t, (*bwcli.PasswordManagerClient)(nil), pm)
}

func TestProviderAuthUsingAccessToken(t *testing.T) {
	cfg := providerConfig{
		AccessToken:          "0.client_id.client_secret:dGVzdC1lbmNyeXB0aW9uLWtleQ==",
		ClientImplementation: schema_definition.ClientImplementationEmbedded,
	}
	assert.NoError(t, validateProviderConfig(cfg))

	clients, err := configureClients(t.Context(), versionTestSkippedLogin, cfg)
	assert.NoError(t, err)

	sm, err := clients.RequireSecretsManager()
	assert.NoError(t, err)
	assert.Implements(t, (*embedded.SecretsManager)(nil), sm)
}

func TestProviderAuthUsingAPIKey_ThrowsErrorOnMissingClientID(t *testing.T) {
	cfg := providerConfig{
		Server:         "http://127.0.0.1/",
		ClientSecret:   "client-secret-5678",
		MasterPassword: "master-password-9",
	}

	err := validateProviderConfig(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "client_id")
	}
}

func TestProviderAuthUsingAPIKey_ThrowsErrorOnMissingClientSecret(t *testing.T) {
	cfg := providerConfig{
		Server:         "http://127.0.0.1/",
		ClientID:       "client-id-1234",
		MasterPassword: "master-password-9",
	}

	err := validateProviderConfig(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "client_secret")
	}
}

func TestProviderAuthUsingAPIKey_ThrowsErrorOnMissingMasterPassword(t *testing.T) {
	cfg := providerConfig{
		Server:       "http://127.0.0.1/",
		Email:        "test@laverse.net",
		ClientID:     "client-id-1234",
		ClientSecret: "client-secret-5678",
	}

	err := validateProviderConfig(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "master_password")
	}
}

func TestProviderAuthUsingUsernamePassword_ThrowsErrorOnMissingMasterPassword(t *testing.T) {
	cfg := providerConfig{
		Server: "http://127.0.0.1/",
		Email:  "test@laverse.net",
	}

	err := validateProviderConfig(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "master_password")
		assert.Contains(t, err.Error(), "session_key")
	}
}

func TestProviderAuthForPasswordManager_ThrowsErrorOnMissingEmail(t *testing.T) {
	cfg := providerConfig{
		Server:         "http://127.0.0.1/",
		MasterPassword: "master-password-9",
	}

	err := validateProviderConfig(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "email")
	}
}

func TestSyncAfterWriteVerificationDisabled(t *testing.T) {
	cfg := providerConfig{
		Server:               "http://127.0.0.1/",
		Email:                "test@laverse.net",
		ClientID:             "client-id-1234",
		ClientSecret:         "client-secret-5678",
		MasterPassword:       "master-password-9",
		ClientImplementation: schema_definition.ClientImplementationEmbedded,
		ExperimentalDisableSyncAfterWriteVerification: true,
	}
	assert.NoError(t, validateProviderConfig(cfg))

	clients, err := configureClients(t.Context(), versionTestSkippedLogin, cfg)
	assert.NoError(t, err)

	pm, err := clients.RequirePasswordManager()
	assert.NoError(t, err)
	assert.Implements(t, (*embedded.PasswordManagerClient)(nil), pm)
	assert.True(t, pm.(embedded.PasswordManagerClient).IsSyncAfterWriteVerificationDisabled())
}

func TestProviderAuthUsingExperimentalEmbeddedClient_BackwardCompatibility(t *testing.T) {
	cfg := providerConfig{
		Server:                     "http://127.0.0.1/",
		Email:                      "test@laverse.net",
		ClientID:                   "client-id-1234",
		ClientSecret:               "client-secret-5678",
		MasterPassword:             "master-password-9",
		ExperimentalEmbeddedClient: true,
	}
	assert.NoError(t, validateProviderConfig(cfg))

	clients, err := configureClients(t.Context(), versionTestSkippedLogin, cfg)
	assert.NoError(t, err)

	pm, err := clients.RequirePasswordManager()
	assert.NoError(t, err)
	assert.Implements(t, (*embedded.PasswordManagerClient)(nil), pm)
}

func TestGetClientImplementation_RecognizesExperimentalEmbeddedClient(t *testing.T) {
	cfg := providerConfig{
		Server:                     "http://127.0.0.1/",
		Email:                      "test@laverse.net",
		ClientID:                   "client-id-1234",
		ClientSecret:               "client-secret-5678",
		MasterPassword:             "master-password-9",
		ExperimentalEmbeddedClient: true,
	}

	clientImpl := getClientImplementation(cfg)
	assert.Equal(t, schema_definition.ClientImplementationEmbedded, clientImpl)
}

func TestGetClientImplementation_RecognizesExplicitClientImplementation(t *testing.T) {
	cfg := providerConfig{
		Server:               "http://127.0.0.1/",
		Email:                "test@laverse.net",
		ClientID:             "client-id-1234",
		ClientSecret:         "client-secret-5678",
		MasterPassword:       "master-password-9",
		ClientImplementation: schema_definition.ClientImplementationEmbedded,
	}

	clientImpl := getClientImplementation(cfg)
	assert.Equal(t, schema_definition.ClientImplementationEmbedded, clientImpl)
}

func TestGetClientImplementation_DefaultsToCLI(t *testing.T) {
	cfg := providerConfig{
		Server:         "http://127.0.0.1/",
		Email:          "test@laverse.net",
		ClientID:       "client-id-1234",
		ClientSecret:   "client-secret-5678",
		MasterPassword: "master-password-9",
		// ClientImplementation not set, should default to "cli"
	}

	clientImpl := getClientImplementation(cfg)
	assert.Equal(t, schema_definition.ClientImplementationCLI, clientImpl)
}

func TestGetClientImplementation_ExperimentalTakesPrecedenceWhenBothSet(t *testing.T) {
	cfg := providerConfig{
		Server:                     "http://127.0.0.1/",
		Email:                      "test@laverse.net",
		ClientID:                   "client-id-1234",
		ClientSecret:               "client-secret-5678",
		MasterPassword:             "master-password-9",
		ClientImplementation:       schema_definition.ClientImplementationCLI, // explicitly set to "cli"
		ExperimentalEmbeddedClient: true,                                      // but experimental.embedded_client is also set
	}

	// Verify that experimental.embedded_client takes precedence over explicit client_implementation
	clientImpl := getClientImplementation(cfg)
	assert.Equal(t, schema_definition.ClientImplementationEmbedded, clientImpl, "experimental.embedded_client should take precedence when both are set")
}
