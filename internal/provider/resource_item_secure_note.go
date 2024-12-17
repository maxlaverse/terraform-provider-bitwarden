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
		CreateContext: withPasswordManager(resourceCreateObject(models.ObjectTypeItem, models.ItemTypeSecureNote)),
		ReadContext:   withPasswordManager(resourceReadObjectIgnoreMissing),
		UpdateContext: withPasswordManager(resourceUpdateObject),
		DeleteContext: withPasswordManager(resourceDeleteObject),
		Importer:      resourceImporter(resourceImportObject(models.ObjectTypeItem, models.ItemTypeSecureNote)),
		Schema:        dataSourceItemSecureNoteSchema,
	}
}
