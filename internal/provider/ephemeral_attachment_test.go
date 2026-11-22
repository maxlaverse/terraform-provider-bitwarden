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

func TestAttachmentReadValidation(t *testing.T) {
	server := NewProviderServer("", WithSkipLogin())()
	schemaResp, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schemaResp.Diagnostics)

	attachmentSchema := schemaResp.EphemeralResourceSchemas["bitwarden_attachment"]
	require.NotNil(t, attachmentSchema, "ephemeral attachment must be registered")

	for _, attr := range attachmentSchema.Block.Attributes {
		if attr.Name == "content" {
			assert.True(t, attr.Sensitive, "content must be sensitive")
		}
	}

	valid := map[string]any{"id": "attachment-id", "item_id": "item-id"}
	tests := []struct {
		name    string
		attrs   map[string]any
		wantErr bool
	}{
		{name: "by ID", attrs: valid},
		{name: "unknown ID", attrs: map[string]any{"id": tftypes.UnknownValue, "item_id": "item-id"}},
		{name: "unknown item ID", attrs: map[string]any{"id": "attachment-id", "item_id": tftypes.UnknownValue}},
		{name: "missing ID", attrs: map[string]any{"item_id": "item-id"}, wantErr: true},
		{name: "missing item ID", attrs: map[string]any{"id": "attachment-id"}, wantErr: true},
		{name: "configured content", attrs: map[string]any{"id": "attachment-id", "item_id": "item-id", "content": "unexpected"}, wantErr: true},
	}

	for _, tt := range tests {
		for _, kind := range []string{"ephemeral", "data source"} {
			t.Run(kind+"/"+tt.name, func(t *testing.T) {
				var diagnostics []*tfprotov6.Diagnostic
				if kind == "ephemeral" {
					resp, err := server.ValidateEphemeralResourceConfig(t.Context(), &tfprotov6.ValidateEphemeralResourceConfigRequest{
						TypeName: "bitwarden_attachment",
						Config:   ephemeralAttachmentConfig(t, tt.attrs),
					})
					require.NoError(t, err)
					diagnostics = resp.Diagnostics
				} else {
					objType := schema_definition.AttachmentDataSourceSchema().Type().TerraformType(t.Context()).(tftypes.Object)
					resp, err := server.ValidateDataResourceConfig(t.Context(), &tfprotov6.ValidateDataResourceConfigRequest{
						TypeName: "bitwarden_attachment",
						Config:   attachmentReadConfig(t, objType, tt.attrs),
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

func TestEphemeralAttachmentProviderClients(t *testing.T) {
	t.Chdir(t.TempDir())

	for _, key := range []string{"BW_PASSWORD", "BW_SESSION", "BW_CLIENTID", "BW_CLIENTSECRET", "NODE_EXTRA_CA_CERTS"} {
		t.Setenv(key, "")
	}

	tests := []struct {
		name           string
		implementation string
		sessionKey     string
		email          string
		masterPassword string
	}{
		{name: "cli", implementation: "cli", email: "test@example.com", sessionKey: t.Name()},
		{name: "embedded", implementation: "embedded", email: "test@example.com", masterPassword: t.Name()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New("", WithSkipLogin())()
			var schemaResp fwprovider.SchemaResponse
			p.Schema(t.Context(), fwprovider.SchemaRequest{}, &schemaResp)

			objType := schemaResp.Schema.Type().TerraformType(t.Context()).(tftypes.Object)
			attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
			for name, typ := range objType.AttributeTypes {
				attrs[name] = tftypes.NewValue(typ, nil)
			}
			attrs["email"] = tftypes.NewValue(tftypes.String, tt.email)
			attrs["server"] = tftypes.NewValue(tftypes.String, "http://127.0.0.1")
			attrs["client_implementation"] = tftypes.NewValue(tftypes.String, tt.implementation)
			attrs["vault_path"] = tftypes.NewValue(tftypes.String, t.TempDir())
			if tt.sessionKey != "" {
				attrs["session_key"] = tftypes.NewValue(tftypes.String, tt.sessionKey)
			}
			if tt.masterPassword != "" {
				attrs["master_password"] = tftypes.NewValue(tftypes.String, tt.masterPassword)
			}

			var resp fwprovider.ConfigureResponse
			p.Configure(t.Context(), fwprovider.ConfigureRequest{
				Config: tfsdk.Config{Raw: tftypes.NewValue(objType, attrs), Schema: schemaResp.Schema},
			}, &resp)
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			clients, ok := resp.EphemeralResourceData.(*ProviderClients)
			require.True(t, ok)
			require.NotNil(t, clients.PasswordManager)
		})
	}
}

func ephemeralAttachmentConfig(t *testing.T, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()

	objType := schema_definition.AttachmentEphemeralResourceSchema().Type().TerraformType(t.Context()).(tftypes.Object)
	return attachmentReadConfig(t, objType, values)
}

func attachmentReadConfig(t *testing.T, objType tftypes.Object, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(typ, values[name])
	}

	config, err := tfprotov6.NewDynamicValue(objType, tftypes.NewValue(objType, attrs))
	require.NoError(t, err)

	return &config
}
