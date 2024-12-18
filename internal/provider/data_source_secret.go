package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceSecret() *schema.Resource {
	dataSourceSecretSchema := schema_definition.SecretSchema(schema_definition.DataSource)

	return &schema.Resource{
		Description: "Use this data source to get information on an existing secret.",
		ReadContext: withSecretsManager(opSecretRead),
		Schema:      dataSourceSecretSchema,
	}
}
