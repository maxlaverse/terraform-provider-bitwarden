//go:build offline

package provider

import (
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretReadValidation(t *testing.T) {
	server := NewProviderServer(versionTestSkippedLogin)()
	schemaResp, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schemaResp.Diagnostics)

	secretSchema := schemaResp.EphemeralResourceSchemas["bitwarden_secret"]
	require.NotNil(t, secretSchema, "ephemeral secret must be registered")

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
		for _, kind := range []string{"ephemeral", "data source"} {
			t.Run(kind+"/"+tt.name, func(t *testing.T) {
				var diagnostics []*tfprotov6.Diagnostic
				if kind == "ephemeral" {
					resp, err := server.ValidateEphemeralResourceConfig(t.Context(), &tfprotov6.ValidateEphemeralResourceConfigRequest{
						TypeName: "bitwarden_secret",
						Config:   ephemeralSecretConfig(t, tt.attrs),
					})
					require.NoError(t, err)
					diagnostics = resp.Diagnostics
				} else {
					objType := schema_definition.SecretDataSourceSchema().Type().TerraformType(t.Context()).(tftypes.Object)
					resp, err := server.ValidateDataResourceConfig(t.Context(), &tfprotov6.ValidateDataResourceConfigRequest{
						TypeName: "bitwarden_secret",
						Config:   secretReadConfig(t, objType, tt.attrs),
					})
					require.NoError(t, err)
					diagnostics = resp.Diagnostics
				}

				if tt.wantErr {
					require.NotEmpty(t, diagnostics)
					assert.Equal(t, tfprotov6.DiagnosticSeverityError, diagnostics[0].Severity)
				} else {
					assert.Empty(t, diagnostics)
				}
			})
		}
	}
}

func TestEphemeralSecretProviderClients(t *testing.T) {
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
		})
	}
}

func ephemeralSecretConfig(t *testing.T, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()

	objType := schema_definition.SecretEphemeralResourceSchema().Type().TerraformType(t.Context()).(tftypes.Object)
	return secretReadConfig(t, objType, values)
}

func secretReadConfig(t *testing.T, objType tftypes.Object, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(typ, values[name])
	}

	config, err := tfprotov6.NewDynamicValue(objType, tftypes.NewValue(objType, attrs))
	require.NoError(t, err)

	return &config
}
