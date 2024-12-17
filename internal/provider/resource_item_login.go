package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceItemLogin() *schema.Resource {
	dataSourceItemSecureNoteSchema := schema_definition.BaseSchema(schema_definition.Resource)
	for k, v := range schema_definition.LoginSchema(schema_definition.Resource) {
		dataSourceItemSecureNoteSchema[k] = v
	}

	return &schema.Resource{
		Description:   "Manages a login item.",
		CreateContext: withPasswordManager(resourceCreateObject(models.ObjectTypeItem, models.ItemTypeLogin)),
		ReadContext:   withPasswordManager(resourceReadObjectIgnoreMissing),
		UpdateContext: withPasswordManager(resourceUpdateObject),
		DeleteContext: withPasswordManager(resourceDeleteObject),
		Importer:      resourceImporter(resourceImportObject(models.ObjectTypeItem, models.ItemTypeLogin)),
		Schema:        dataSourceItemSecureNoteSchema,
	}
}
