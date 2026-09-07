//go:build offline

package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEphemeralSecretValidation(t *testing.T) {
	factory, err := NewProviderServer(versionTestSkippedLogin)
	require.NoError(t, err)

	server := factory()
	schemaResp, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schemaResp.Diagnostics)

	secretSchema := schemaResp.EphemeralResourceSchemas["bitwarden_secret"]
	require.NotNil(t, secretSchema, "ephemeral secret must be registered in the mux")

	for _, attr := range secretSchema.Block.Attributes {
		if attr.Name == "value" || attr.Name == "note" {
			assert.True(t, attr.Sensitive, "%s must be sensitive", attr.Name)
		}
	}

	tests := []struct {
		name    string
		attrs   map[string]any
		wantErr bool
	}{
		{name: "by ID", attrs: map[string]any{"id": "secret-id"}},
		{name: "by key", attrs: map[string]any{"key": "secret-key"}},
		{name: "unknown ID", attrs: map[string]any{"id": tftypes.UnknownValue}},
		{name: "unknown key", attrs: map[string]any{"key": tftypes.UnknownValue}},
		{name: "missing selector", wantErr: true},
		{name: "both selectors", attrs: map[string]any{"id": "secret-id", "key": "secret-key"}, wantErr: true},
		{name: "empty ID", attrs: map[string]any{"id": ""}, wantErr: true},
		{name: "empty key", attrs: map[string]any{"key": ""}, wantErr: true},
		{name: "configured value", attrs: map[string]any{"id": "secret-id", "value": "unexpected"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.ValidateEphemeralResourceConfig(t.Context(), &tfprotov6.ValidateEphemeralResourceConfigRequest{
				TypeName: "bitwarden_secret",
				Config:   ephemeralSecretConfig(t, tt.attrs),
			})
			require.NoError(t, err)

			if tt.wantErr {
				require.NotEmpty(t, resp.Diagnostics)
				assert.Equal(t, tfprotov6.DiagnosticSeverityError, resp.Diagnostics[0].Severity)
			} else {
				assert.Empty(t, resp.Diagnostics)
			}
		})
	}
}

func TestEphemeralSecretOpenAndClose(t *testing.T) {
	secret := models.Secret{
		ID:             "secret-id",
		Key:            "secret-key",
		Value:          "private-value",
		Note:           "private-note",
		OrganizationID: "organization-id",
		ProjectID:      "project-id",
	}

	for _, selector := range []string{"id", "key"} {
		t.Run(selector, func(t *testing.T) {
			client := &ephemeralSecretClient{secret: &secret}
			server := ephemeralSecretServer(t, &ProviderClients{SecretsManager: client})

			input := secret.ID
			if selector == "key" {
				input = secret.Key
			}

			resp, err := server.OpenEphemeralResource(t.Context(), &tfprotov6.OpenEphemeralResourceRequest{
				TypeName: "bitwarden_secret",
				Config:   ephemeralSecretConfig(t, map[string]any{selector: input}),
			})
			require.NoError(t, err)
			require.Empty(t, resp.Diagnostics)
			require.NotNil(t, resp.Result)

			result, err := resp.Result.Unmarshal(schema_definition.SecretEphemeralResourceSchema().Type().TerraformType(t.Context()))
			require.NoError(t, err)

			var attrs map[string]tftypes.Value
			require.NoError(t, result.As(&attrs))

			for name, want := range map[string]string{
				"id": secret.ID, "key": secret.Key, "value": secret.Value,
				"note": secret.Note, "organization_id": secret.OrganizationID, "project_id": secret.ProjectID,
			} {
				var got string
				require.NoError(t, attrs[name].As(&got))
				assert.Equal(t, want, got, name)
			}

			assert.Equal(t, 1, client.reads)
			assert.Equal(t, selector, client.selector)
			assert.Equal(t, input, client.input)
			assert.Empty(t, resp.Private, "no decrypted data should be kept for Close")
			assert.True(t, resp.RenewAt.IsZero(), "reading an existing secret does not create a lease")

			closed, err := server.CloseEphemeralResource(t.Context(), &tfprotov6.CloseEphemeralResourceRequest{
				TypeName: "bitwarden_secret",
				Private:  resp.Private,
			})
			require.NoError(t, err)
			assert.Empty(t, closed.Diagnostics)
			// Mutation methods on the stub are intentionally unimplemented:
			// attempting to delete the stored secret would panic this test.
			assert.Equal(t, 1, client.reads)
		})
	}
}

func TestEphemeralSecretOpenErrors(t *testing.T) {
	tests := []struct {
		name   string
		client bitwarden.SecretsManager
		want   string
	}{
		{name: "no Secrets Manager client", want: "Provider not configured for Secrets Manager"},
		{name: "not found", client: &ephemeralSecretClient{err: models.ErrObjectNotFound}, want: "object not found"},
		{name: "ambiguous key", client: &ephemeralSecretClient{err: models.ErrTooManyObjectsFound}, want: "too many objects found"},
		{name: "canceled", client: &ephemeralSecretClient{err: context.Canceled}, want: "context canceled"},
		{name: "raw CLI output", client: &ephemeralSecretClient{err: errors.New("invalid JSON: private-value")}, want: "Client error details are omitted"},
		{name: "wrapped error", client: &ephemeralSecretClient{err: fmt.Errorf("private-value: %w", models.ErrObjectNotFound)}, want: "object not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := ephemeralSecretServer(t, &ProviderClients{SecretsManager: tt.client})
			resp, err := server.OpenEphemeralResource(t.Context(), &tfprotov6.OpenEphemeralResourceRequest{
				TypeName: "bitwarden_secret",
				Config:   ephemeralSecretConfig(t, map[string]any{"key": "secret-key"}),
			})
			require.NoError(t, err)
			require.Len(t, resp.Diagnostics, 1)

			diagnostic := resp.Diagnostics[0]
			assert.Equal(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity)
			assert.Contains(t, diagnostic.Summary+diagnostic.Detail, tt.want)
			assert.NotContains(t, diagnostic.Summary+diagnostic.Detail, "private-value")
			assert.Empty(t, resp.Private)
		})
	}
}

func TestEphemeralSecretConfigure(t *testing.T) {
	for _, tt := range []struct {
		name string
		data any
	}{
		{name: "not configured yet"},
		{name: "wrong provider data", data: "wrong type"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &secretEphemeralResource{}
			var resp ephemeral.ConfigureResponse

			r.Configure(t.Context(), ephemeral.ConfigureRequest{ProviderData: tt.data}, &resp)

			assert.Equal(t, tt.data != nil, resp.Diagnostics.HasError())
			assert.Nil(t, r.clients)
		})
	}
}

func TestEphemeralSecretSharedClients(t *testing.T) {
	t.Chdir(t.TempDir())

	for _, key := range []string{"BW_PASSWORD", "BW_SESSION", "BW_CLIENTID", "BW_CLIENTSECRET", "NODE_EXTRA_CA_CERTS"} {
		t.Setenv(key, "")
	}

	for _, implementation := range []string{"cli", "embedded"} {
		t.Run(implementation, func(t *testing.T) {
			p := New(versionTestSkippedLogin)()
			var schemaResp fwprovider.SchemaResponse
			p.Schema(t.Context(), fwprovider.SchemaRequest{}, &schemaResp)

			objType := schemaResp.Schema.Type().TerraformType(t.Context()).(tftypes.Object)
			attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
			for name, typ := range objType.AttributeTypes {
				attrs[name] = tftypes.NewValue(typ, nil)
			}
			attrs["access_token"] = tftypes.NewValue(tftypes.String, t.Name())
			attrs["server"] = tftypes.NewValue(tftypes.String, "http://127.0.0.1")
			attrs["client_implementation"] = tftypes.NewValue(tftypes.String, implementation)
			attrs["vault_path"] = tftypes.NewValue(tftypes.String, t.TempDir())

			var resp fwprovider.ConfigureResponse
			p.Configure(t.Context(), fwprovider.ConfigureRequest{
				Config: tfsdk.Config{Raw: tftypes.NewValue(objType, attrs), Schema: schemaResp.Schema},
			}, &resp)
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			clients, ok := resp.EphemeralResourceData.(*ProviderClients)
			require.True(t, ok)
			require.NotNil(t, clients.SecretsManager)
			t.Cleanup(func() {
				muxClientsMu.Lock()
				defer muxClientsMu.Unlock()

				for key, offered := range muxClientsOffer {
					if offered == clients {
						delete(muxClientsOffer, key)
					}
				}
			})
			assert.Same(t, resp.ResourceData, clients)
			assert.Same(t, resp.DataSourceData, clients)

			r := &secretEphemeralResource{}
			var configured ephemeral.ConfigureResponse
			r.Configure(t.Context(), ephemeral.ConfigureRequest{ProviderData: clients}, &configured)
			require.False(t, configured.Diagnostics.HasError(), configured.Diagnostics)
			assert.Same(t, clients, r.clients)
		})
	}
}

type ephemeralSecretClient struct {
	bitwarden.SecretsManager
	secret   *models.Secret
	err      error
	reads    int
	selector string
	input    string
}

func (c *ephemeralSecretClient) GetSecret(_ context.Context, secret models.Secret) (*models.Secret, error) {
	c.reads++
	c.selector = "id"
	c.input = secret.ID
	return c.secret, c.err
}

func (c *ephemeralSecretClient) GetSecretByKey(_ context.Context, key string) (*models.Secret, error) {
	c.reads++
	c.selector = "key"
	c.input = key
	return c.secret, c.err
}

type ephemeralSecretTestProvider struct {
	*bitwardenProvider
	clients *ProviderClients
}

func (p *ephemeralSecretTestProvider) Configure(_ context.Context, _ fwprovider.ConfigureRequest, resp *fwprovider.ConfigureResponse) {
	resp.EphemeralResourceData = p.clients
}

func ephemeralSecretServer(t *testing.T, clients *ProviderClients) tfprotov6.ProviderServer {
	t.Helper()

	server := providerserver.NewProtocol6(&ephemeralSecretTestProvider{
		bitwardenProvider: &bitwardenProvider{},
		clients:           clients,
	})()

	resp, err := server.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{})
	require.NoError(t, err)
	require.Empty(t, resp.Diagnostics)

	return server
}

func ephemeralSecretConfig(t *testing.T, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()

	objType := schema_definition.SecretEphemeralResourceSchema().Type().TerraformType(t.Context()).(tftypes.Object)
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(typ, values[name])
	}

	config, err := tfprotov6.NewDynamicValue(objType, tftypes.NewValue(objType, attrs))
	require.NoError(t, err)

	return &config
}
