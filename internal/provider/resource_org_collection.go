package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceOrgCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an organization collection.",

		CreateContext: withPasswordManager(opOrganizationCollectionCreate),
		ReadContext:   withPasswordManager(opObjectReadIgnoreMissing),
		UpdateContext: withPasswordManager(opObjectUpdate),
		DeleteContext: withPasswordManager(opObjectDelete),
		Importer:      resourceImporter(opOrganizationCollectionImport),

		Schema: schema_definition.OrgCollectionSchema(schema_definition.Resource),
	}
}
