package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func resourceItemLogin() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(Resource)
	for k, v := range loginSchema(Resource) {
		dataSourceItemSecureNoteSchema[k] = v
	}

	return &schema.Resource{
		Description:   "Use this resource to set (amongst other things) the username and password of a Bitwarden Login item.",
		CreateContext: createResource(bw.ObjectTypeItem, bw.ItemTypeLogin),
		ReadContext:   objectRead,
		UpdateContext: objectUpdate,
		DeleteContext: objectDelete,
		Importer:      importItemResource(bw.ObjectTypeItem, bw.ItemTypeLogin),
		Schema:        dataSourceItemSecureNoteSchema,
	}
}
