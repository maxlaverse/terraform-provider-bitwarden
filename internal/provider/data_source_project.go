package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceProject() *schema.Resource {
	dataSourceProjectSchema := schema_definition.ProjectSchema(schema_definition.DataSource)

	return &schema.Resource{
		Description: "Use this data source to get information on an existing project.",
		ReadContext: withSecretsManager(opProjectRead),
		Schema:      dataSourceProjectSchema,
	}
}
