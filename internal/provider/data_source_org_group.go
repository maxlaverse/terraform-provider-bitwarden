package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceOrgGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing organization group.",
		ReadContext: withPasswordManager(opOrganizationGroupRead),
		Schema:      schema_definition.OrgGroupSchema(schema_definition.DataSource),
	}
}
