package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceFolder() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a folder.",

		CreateContext: withPasswordManager(opFolderCreate),
		ReadContext:   withPasswordManager(opObjectReadIgnoreMissing),
		UpdateContext: withPasswordManager(opObjectUpdate),
		DeleteContext: withPasswordManager(opObjectDelete),
		Importer:      resourceImporter(opFolderImport),

		Schema: schema_definition.FolderSchema(schema_definition.Resource),
	}
}
