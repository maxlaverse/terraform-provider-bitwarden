package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceOrgMember() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an organization member.",

		CreateContext: withPasswordManager(opOrganizationMemberCreate),
		ReadContext:   withPasswordManager(opOrganizationMemberReadIgnoreMissing),
		DeleteContext: withPasswordManager(opOrganizationMemberDelete),
		Importer:      resourceImporter(opOrganizationMemberImport),

		Schema: schema_definition.OrgMemberSchema(schema_definition.Resource),
	}
}
