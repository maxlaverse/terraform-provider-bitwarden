//go:build offline

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemSecureNoteReadValidation(t *testing.T) {
	server := NewProviderServer("", WithSkipLogin())()
	schemaResp, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schemaResp.Diagnostics)

	noteSchema := schemaResp.EphemeralResourceSchemas["bitwarden_item_secure_note"]
	require.NotNil(t, noteSchema, "ephemeral secure note item must be registered")

	for _, attr := range noteSchema.Block.Attributes {
		switch attr.Name {
		case "notes", "field":
			assert.True(t, attr.Sensitive, "%s must be sensitive", attr.Name)
		}
	}

	tests := []struct {
		name    string
		attrs   map[string]any
		wantErr bool
	}{
		{name: "by ID", attrs: map[string]any{"id": "item-id"}},
		{name: "by search", attrs: map[string]any{"search": "VPN/Pre-sharedSecret"}},
		{name: "unknown ID", attrs: map[string]any{"id": tftypes.UnknownValue}},
		{name: "unknown search", attrs: map[string]any{"search": tftypes.UnknownValue}},
		{name: "both selectors", attrs: map[string]any{"id": "item-id", "search": "VPN/Pre-sharedSecret"}},
		{name: "missing selector", wantErr: true},
		{name: "configured notes", attrs: map[string]any{"id": "item-id", "notes": "unexpected"}, wantErr: true},
	}

	for _, tt := range tests {
		for _, kind := range []string{"ephemeral", "data source"} {
			t.Run(kind+"/"+tt.name, func(t *testing.T) {
				var diagnostics []*tfprotov6.Diagnostic
				if kind == "ephemeral" {
					resp, err := server.ValidateEphemeralResourceConfig(t.Context(), &tfprotov6.ValidateEphemeralResourceConfigRequest{
						TypeName: "bitwarden_item_secure_note",
						Config:   ephemeralItemSecureNoteConfig(t, tt.attrs),
					})
					require.NoError(t, err)
					diagnostics = resp.Diagnostics
				} else {
					objType := schema_definition.SecureNoteDataSourceSchema().Type().TerraformType(t.Context()).(tftypes.Object)
					resp, err := server.ValidateDataResourceConfig(t.Context(), &tfprotov6.ValidateDataResourceConfigRequest{
						TypeName: "bitwarden_item_secure_note",
						Config:   itemSecureNoteReadConfig(t, objType, tt.attrs),
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

func ephemeralItemSecureNoteConfig(t *testing.T, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()

	objType := schema_definition.SecureNoteEphemeralResourceSchema().Type().TerraformType(t.Context()).(tftypes.Object)
	return itemSecureNoteReadConfig(t, objType, values)
}

func itemSecureNoteReadConfig(t *testing.T, objType tftypes.Object, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(typ, values[name])
	}

	config, err := tfprotov6.NewDynamicValue(objType, tftypes.NewValue(objType, attrs))
	require.NoError(t, err)

	return &config
}
