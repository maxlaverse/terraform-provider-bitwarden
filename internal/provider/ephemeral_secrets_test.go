//go:build offline

package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEphemeralSecretsValidation(t *testing.T) {
	server := NewProviderServer(versionTestSkippedLogin)()
	schema, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.Empty(t, schema.Diagnostics)

	batchSchema := schema.EphemeralResourceSchemas["bitwarden_secrets"]
	require.NotNil(t, batchSchema)
	assert.NotContains(t, schema.ResourceSchemas, "bitwarden_secrets")
	assert.NotContains(t, schema.DataSourceSchemas, "bitwarden_secrets")

	for _, attr := range batchSchema.Block.Attributes {
		if attr.Name == "ids" {
			assert.True(t, attr.Required)
			assert.True(t, attr.Type.Equal(tftypes.Set{ElementType: tftypes.String}))
		}
		if attr.Name == "secrets" {
			assert.True(t, attr.Sensitive)
			assert.True(t, attr.Computed)
		}
	}

	tests := []struct {
		name    string
		ids     any
		values  map[string]string
		wantErr bool
	}{
		{name: "IDs", ids: []string{"first-id", "second-id"}},
		{name: "case-insensitive IDs", ids: []string{"secret-id", "SECRET-ID"}},
		{name: "empty set", ids: []string{}},
		{name: "unknown set", ids: tftypes.UnknownValue},
		{name: "unknown ID", ids: []tftypes.Value{tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}},
		{name: "missing set", wantErr: true},
		{name: "empty ID", ids: []string{""}, wantErr: true},
		{name: "null ID", ids: []tftypes.Value{tftypes.NewValue(tftypes.String, nil)}, wantErr: true},
		{name: "configured secrets", ids: []string{}, values: map[string]string{"secret-id": "private-value"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.ValidateEphemeralResourceConfig(t.Context(), &tfprotov6.ValidateEphemeralResourceConfigRequest{
				TypeName: "bitwarden_secrets",
				Config:   ephemeralSecretsConfig(t, tt.ids, tt.values),
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

func TestEphemeralSecretsProviderConfiguration(t *testing.T) {
	testEphemeralSecretProviderConfiguration(t, "bitwarden_secrets", `ids = ["secret-id"]`, `ephemeral.bitwarden_secrets.credential.secrets["secret-id"].value`)
}

func (c *ephemeralCredentialClient) GetSecretsByIDs(ctx context.Context, ids []string) ([]models.Secret, error) {
	secrets := make([]models.Secret, 0, len(ids))
	for _, id := range ids {
		secret, err := c.GetSecret(ctx, models.Secret{ID: id})
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, *secret)
	}
	return secrets, nil
}

func TestEphemeralSecretsOpenCloseAndReopen(t *testing.T) {
	client := &ephemeralSecretsClient{secrets: []models.Secret{{ID: "secret-id", Value: "first-private-value"}}}
	server := ephemeralSecretServer(t, &ProviderClients{SecretsManager: client})
	ids := []string{"secret-id", "SECRET-ID"}

	for phase := range 2 {
		value := fmt.Sprintf("private-value-%d\"\n雪", phase)
		client.secrets = []models.Secret{{ID: "secret-id", Key: "secret-key", Value: value, Note: "private-note"}}

		resp, err := server.OpenEphemeralResource(t.Context(), &tfprotov6.OpenEphemeralResourceRequest{
			TypeName: "bitwarden_secrets",
			Config:   ephemeralSecretsConfig(t, ids, nil),
		})
		require.NoError(t, err)
		require.Empty(t, resp.Diagnostics)
		assert.Equal(t, map[string]string{"secret-id": value}, ephemeralSecretsValues(t, resp))
		assert.Equal(t, map[string]batchSecretModel{"secret-id": {Key: "secret-key", Value: value, Note: "private-note"}}, ephemeralSecretsRecords(t, resp))
		assert.Equal(t, phase+1, client.reads, "each phase must fetch fresh values")
		assert.Empty(t, resp.Private)
		assert.True(t, resp.RenewAt.IsZero())

		closed, err := server.CloseEphemeralResource(t.Context(), &tfprotov6.CloseEphemeralResourceRequest{
			TypeName: "bitwarden_secrets",
			Private:  resp.Private,
		})
		require.NoError(t, err)
		assert.Empty(t, closed.Diagnostics)
		assert.Equal(t, phase+1, client.reads)
	}

	resp, err := server.OpenEphemeralResource(t.Context(), &tfprotov6.OpenEphemeralResourceRequest{
		TypeName: "bitwarden_secrets",
		Config:   ephemeralSecretsConfig(t, []string{}, nil),
	})
	require.NoError(t, err)
	require.Empty(t, resp.Diagnostics)
	assert.Empty(t, ephemeralSecretsValues(t, resp))
	assert.Equal(t, 2, client.reads, "an empty batch must not call the client")
}

func TestEphemeralSecretsOpenErrors(t *testing.T) {
	tests := []struct {
		name   string
		client bitwarden.SecretsManager
		want   string
	}{
		{name: "not configured", want: "Provider not configured for Secrets Manager"},
		{name: "not found", client: &ephemeralSecretsClient{err: models.ErrObjectNotFound}, want: "object not found"},
		{name: "canceled", client: &ephemeralSecretsClient{err: context.Canceled}, want: "context canceled"},
		{name: "logged out", client: &ephemeralSecretsClient{err: models.ErrLoggedOut}, want: models.ErrLoggedOut.Error()},
		{name: "untrusted error", client: &ephemeralSecretsClient{err: errors.New("private-value")}, want: "Client error details are omitted"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := ephemeralSecretServer(t, &ProviderClients{SecretsManager: tt.client})
			resp, err := server.OpenEphemeralResource(t.Context(), &tfprotov6.OpenEphemeralResourceRequest{
				TypeName: "bitwarden_secrets",
				Config:   ephemeralSecretsConfig(t, []string{"secret-id"}, nil),
			})
			require.NoError(t, err)
			require.Len(t, resp.Diagnostics, 1)

			assert.Contains(t, resp.Diagnostics[0].Summary+resp.Diagnostics[0].Detail, tt.want)
			assert.NotContains(t, fmt.Sprint(resp.Diagnostics), "private-value")
			assertEphemeralSecretsNoValues(t, resp)
			assert.Empty(t, resp.Private)
		})
	}
}

type ephemeralSecretsClient struct {
	bitwarden.SecretsManager
	secrets []models.Secret
	err     error
	reads   int
}

func (c *ephemeralSecretsClient) GetSecretsByIDs(_ context.Context, _ []string) ([]models.Secret, error) {
	c.reads++
	return append([]models.Secret(nil), c.secrets...), c.err
}

func ephemeralSecretsConfig(t *testing.T, ids any, values map[string]string) *tfprotov6.DynamicValue {
	t.Helper()
	objType := schema_definition.SecretsEphemeralResourceSchema().Type().TerraformType(t.Context()).(tftypes.Object)

	if input, ok := ids.([]string); ok {
		elements := make([]tftypes.Value, len(input))
		for i, id := range input {
			elements[i] = tftypes.NewValue(tftypes.String, id)
		}
		ids = elements
	}
	var valueInput any
	if values != nil {
		elementType := objType.AttributeTypes["secrets"].(tftypes.Map).ElementType
		valueMap := make(map[string]tftypes.Value, len(values))
		for id, value := range values {
			valueMap[id] = tftypes.NewValue(elementType, map[string]tftypes.Value{
				"key":   tftypes.NewValue(tftypes.String, "secret-key"),
				"value": tftypes.NewValue(tftypes.String, value),
				"note":  tftypes.NewValue(tftypes.String, "private-note"),
			})
		}
		valueInput = valueMap
	}

	config, err := tfprotov6.NewDynamicValue(objType, tftypes.NewValue(objType, map[string]tftypes.Value{
		"ids":     tftypes.NewValue(objType.AttributeTypes["ids"], ids),
		"secrets": tftypes.NewValue(objType.AttributeTypes["secrets"], valueInput),
	}))
	require.NoError(t, err)

	return &config
}

func ephemeralSecretsValues(t *testing.T, resp *tfprotov6.OpenEphemeralResourceResponse) map[string]string {
	t.Helper()
	got := map[string]string{}
	for id, secret := range ephemeralSecretsRecords(t, resp) {
		got[id] = secret.Value
	}
	return got
}

func ephemeralSecretsRecords(t *testing.T, resp *tfprotov6.OpenEphemeralResourceResponse) map[string]batchSecretModel {
	t.Helper()
	require.NotNil(t, resp.Result)

	result, err := resp.Result.Unmarshal(schema_definition.SecretsEphemeralResourceSchema().Type().TerraformType(t.Context()))
	require.NoError(t, err)

	var attrs map[string]tftypes.Value
	require.NoError(t, result.As(&attrs))

	var values map[string]tftypes.Value
	require.NoError(t, attrs["secrets"].As(&values))

	got := make(map[string]batchSecretModel, len(values))
	for id, value := range values {
		var fields map[string]tftypes.Value
		require.NoError(t, value.As(&fields))
		var secret batchSecretModel
		require.NoError(t, fields["key"].As(&secret.Key))
		require.NoError(t, fields["value"].As(&secret.Value))
		require.NoError(t, fields["note"].As(&secret.Note))
		got[id] = secret
	}

	return got
}

func assertEphemeralSecretsNoValues(t *testing.T, resp *tfprotov6.OpenEphemeralResourceResponse) {
	t.Helper()
	if resp.Result == nil {
		return
	}

	// Framework can return the input config on failure, with values still null.
	result, err := resp.Result.Unmarshal(schema_definition.SecretsEphemeralResourceSchema().Type().TerraformType(t.Context()))
	require.NoError(t, err)

	var attrs map[string]tftypes.Value
	require.NoError(t, result.As(&attrs))
	assert.True(t, attrs["secrets"].IsNull(), "errors must not return partial plaintext values")
}
