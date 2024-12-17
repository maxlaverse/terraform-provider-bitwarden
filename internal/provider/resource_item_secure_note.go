package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := schema_definition.BaseSchema(schema_definition.Resource)

	return &schema.Resource{
		Description:   "Manages a secure note item.",
		CreateContext: withPasswordManager(opObjectCreate(models.ObjectTypeItem, models.ItemTypeSecureNote)),
		ReadContext:   withPasswordManager(opObjectReadIgnoreMissing),
		UpdateContext: withPasswordManager(opObjectUpdate),
		DeleteContext: withPasswordManager(opObjectDelete),
		Importer:      resourceImporter(opObjectImport(models.ObjectTypeItem, models.ItemTypeSecureNote)),
		Schema:        dataSourceItemSecureNoteSchema,
	}
}
