package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSecret() *schema.Resource {
	dataSourceSecretSchema := secretSchema(DataSource)

	return &schema.Resource{
		Description: "Use this data source to get information on an existing secret.",
		ReadContext: withSecretsManager(resourceReadDataSourceSecret),
		Schema:      dataSourceSecretSchema,
	}
}
