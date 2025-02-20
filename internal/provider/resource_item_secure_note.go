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
		ReadContext:   withPasswordManager(opItemReadIgnoreMissing(models.ItemTypeSecureNote)),
		UpdateContext: withPasswordManager(opItemUpdate(models.ItemTypeSecureNote)),
		DeleteContext: withPasswordManager(opItemDelete(models.ItemTypeSecureNote)),
		Importer:      resourceImporter(opItemImport),
		Schema:        schema_definition.ItemBaseSchema(schema_definition.Resource),
	}
}
