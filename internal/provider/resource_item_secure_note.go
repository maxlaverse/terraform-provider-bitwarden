package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func resourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(Resource)

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
