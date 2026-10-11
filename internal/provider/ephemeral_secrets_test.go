//go:build offline

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
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

	for _, attr := range batchSchema.Block.Attributes {
		if attr.Name == "ids" {
			assert.True(t, attr.Required)
			assert.True(t, attr.Type.Equal(tftypes.Set{ElementType: tftypes.String}))
		}
		if attr.Name == "secrets" {
			assert.True(t, attr.Sensitive)
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
	secret := models.Secret{
		ID: "secret-id", Key: "credential", Value: ephemeralCredential,
		Note: "ephemeral-note-must-not-be-persisted",
	}
	client := &ephemeralSecretsClient{secrets: []models.Secret{secret}}
	want := map[string]map[string]string{
		secret.ID: {"key": secret.Key, "value": secret.Value, "note": secret.Note},
	}
	check := ephemeralSecretPersistenceCheck{}
	step := resource.TestStep{
		Config:            ephemeralSecretsConsumerConfig(`ids = ["secret-id", "SECRET-ID"]`),
		ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{check}},
		ConfigStateChecks: []statecheck.StateCheck{check},
		Check:             resource.TestCheckOutput("project_name", "public-project"),
	}
	updated := step
	updated.PreConfig = func() {
		client.secrets[0].Value += "-updated"
		want[secret.ID]["value"] = client.secrets[0].Value
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ephemeralSecretsFactories(client, want),
		Steps:                    []resource.TestStep{step, updated},
	})
}

func TestEphemeralSecretsEmpty(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ephemeralSecretsFactories(&ephemeralSecretsClient{err: models.ErrObjectNotFound}, map[string]map[string]string{}),
		Steps: []resource.TestStep{{
			Config: ephemeralSecretsConsumerConfig(`ids = []`),
			Check:  resource.TestCheckOutput("project_name", "public-project"),
		}},
	})
}

func TestEphemeralSecretsErrors(t *testing.T) {
	tests := []struct {
		name   string
		client bitwarden.SecretsManager
		want   string
	}{
		{name: "not configured", want: "Provider not configured for Secrets Manager"},
		{name: "not found", client: &ephemeralSecretsClient{err: models.ErrObjectNotFound}, want: "object not found"},
		{name: "backend error details", client: &ephemeralSecretsClient{err: errors.New("backend unavailable")}, want: "backend unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralSecretsFactories(tt.client, nil),
				Steps: []resource.TestStep{{
					Config:      ephemeralSecretsConsumerConfig(`ids = ["secret-id"]`),
					ExpectError: regexp.MustCompile(tt.want),
				}},
			})
		})
	}
}

type ephemeralSecretsClient struct {
	bitwarden.SecretsManager
	secrets []models.Secret
	err     error
}

func (c *ephemeralSecretsClient) GetSecretsByIDs(_ context.Context, _ []string) ([]models.Secret, error) {
	return append([]models.Secret(nil), c.secrets...), c.err
}

func ephemeralSecretsConsumerConfig(selector string) string {
	return ephemeralCredentialConsumerConfig("bitwarden_secrets", selector, "jsonencode(ephemeral.bitwarden_secrets.credential.secrets)")
}

func ephemeralSecretsFactories(client bitwarden.SecretsManager, want map[string]map[string]string) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"bitwarden": func() (tfprotov6.ProviderServer, error) {
			credential, _ := json.Marshal(want)
			return providerserver.NewProtocol6(&ephemeralCredentialProvider{
				bitwardenProvider: &bitwardenProvider{},
				sourceClients:     &ProviderClients{SecretsManager: client},
				credential:        string(credential),
			})(), nil
		},
	}
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
