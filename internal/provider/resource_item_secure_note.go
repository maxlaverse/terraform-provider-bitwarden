package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func resourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(Resource)

	return &schema.Resource{
		Description:   "Use this resource to set (amongst other things) the content of a Bitwarden Secret Note.",
		CreateContext: createResource(bw.ObjectTypeItem, bw.ItemTypeSecureNote),
		ReadContext:   objectRead,
		UpdateContext: objectUpdate,
		DeleteContext: objectDelete,
		Importer:      importItemResource(bw.ObjectTypeItem, bw.ItemTypeSecureNote),
		Schema:        dataSourceItemSecureNoteSchema,
	}
}
