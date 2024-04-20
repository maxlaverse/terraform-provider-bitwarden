package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func resourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(Resource)

	return &schema.Resource{
		Description:   "Manages a secure note item.",
		CreateContext: createResource(bw.ObjectTypeItem, bw.ItemTypeSecureNote),
		ReadContext:   objectReadIgnoreMissing,
		UpdateContext: objectUpdate,
		DeleteContext: objectDelete,
		Importer:      importItemResource(bw.ObjectTypeItem, bw.ItemTypeSecureNote),
		Schema:        dataSourceItemSecureNoteSchema,
	}
}
