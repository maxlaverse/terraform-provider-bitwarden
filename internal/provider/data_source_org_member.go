package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceOrgMember() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing organization member.",
		ReadContext: withPasswordManager(opOrganizationMemberRead),
		Schema:      schema_definition.OrgMemberSchema(),
	}
}
