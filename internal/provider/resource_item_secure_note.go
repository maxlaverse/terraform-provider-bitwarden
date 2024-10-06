package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func resourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(Resource)

	return &schema.Resource{
		Description:   "Manages a secure note item.",
		CreateContext: resourceCreateObject(models.ObjectTypeItem, models.ItemTypeSecureNote),
		ReadContext:   resourceReadObjectIgnoreMissing,
		UpdateContext: resourceUpdateObject,
		DeleteContext: resourceDeleteObject,
		Importer:      resourceImportObject(models.ObjectTypeItem, models.ItemTypeSecureNote),
		Schema:        dataSourceItemSecureNoteSchema,
	}
}
