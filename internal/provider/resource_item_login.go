package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceItemLogin() *schema.Resource {
	itemLoginSchema := schema_definition.ItemBaseSchema(schema_definition.Resource)
	for k, v := range schema_definition.LoginSchema(schema_definition.Resource) {
		itemLoginSchema[k] = v
	}

	return &schema.Resource{
		Description:   "Manages a login item.",
		CreateContext: withPasswordManager(opItemCreate(models.ItemTypeLogin)),
		ReadContext:   withPasswordManager(opItemReadIgnoreMissing(models.ItemTypeLogin)),
		UpdateContext: withPasswordManager(opItemUpdate(models.ItemTypeLogin)),
		DeleteContext: withPasswordManager(opItemDelete(models.ItemTypeLogin)),
		Importer:      resourceImporter(opItemImport),
		Schema:        itemLoginSchema,
	}
}
