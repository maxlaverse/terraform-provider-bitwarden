package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceProject() *schema.Resource {
	dataSourceProjectSchema := projectSchema(DataSource)

	return &schema.Resource{
		Description: "Use this data source to get information on an existing project.",
		ReadContext: withSecretsManager(resourceReadDataSourceProject),
		Schema:      dataSourceProjectSchema,
	}
}
