package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceOrgGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an organization group.",

		CreateContext: withPasswordManager(opOrganizationGroupCreate),
		ReadContext:   withPasswordManager(opOrganizationGroupReadIgnoreMissing),
		UpdateContext: withPasswordManager(opOrganizationGroupUpdate),
		DeleteContext: withPasswordManager(opOrganizationGroupDelete),
		Importer:      resourceImporter(opOrganizationGroupImport),

		Schema: schema_definition.OrgGroupSchema(schema_definition.Resource),
	}
}
