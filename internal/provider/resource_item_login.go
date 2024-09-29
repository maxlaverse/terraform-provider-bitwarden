package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func resourceItemLogin() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(Resource)
	for k, v := range loginSchema(Resource) {
		dataSourceItemSecureNoteSchema[k] = v
	}

	return &schema.Resource{
		Description:   "Manages a login item.",
		CreateContext: createResource(models.ObjectTypeItem, models.ItemTypeLogin),
		ReadContext:   objectReadIgnoreMissing,
		UpdateContext: objectUpdate,
		DeleteContext: objectDelete,
		Importer:      importItemResource(models.ObjectTypeItem, models.ItemTypeLogin),
		Schema:        dataSourceItemSecureNoteSchema,
	}
}
