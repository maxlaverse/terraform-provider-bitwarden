package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceItemSecureNote() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a secure note item.",
		CreateContext: withPasswordManager(opItemCreate(models.ItemTypeSecureNote)),
		ReadContext:   withPasswordManager(opItemReadIgnoreMissing),
		UpdateContext: withPasswordManager(opItemUpdate),
		DeleteContext: withPasswordManager(opItemDelete),
		Importer:      resourceImporter(opItemImport(models.ItemTypeSecureNote)),
		Schema:        schema_definition.BaseSchema(schema_definition.Resource),
	}
}
