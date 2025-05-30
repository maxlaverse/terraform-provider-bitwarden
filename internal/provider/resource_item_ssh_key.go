package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceItemSSHKey() *schema.Resource {
	itemSSHKeySchema := schema_definition.ItemBaseSchema(schema_definition.Resource)
	for k, v := range schema_definition.SSHKeySchema(schema_definition.Resource) {
		itemSSHKeySchema[k] = v
	}

	return &schema.Resource{
		Description:   "Manages an SSH key item.",
		CreateContext: withPasswordManager(opItemCreate(models.ItemTypeSSHKey)),
		ReadContext:   withPasswordManager(opItemReadIgnoreMissing(models.ItemTypeSSHKey)),
		UpdateContext: withPasswordManager(opItemUpdate(models.ItemTypeSSHKey)),
		DeleteContext: withPasswordManager(opItemDelete(models.ItemTypeSSHKey)),
		Importer:      resourceImporter(opItemImport),
		Schema:        itemSSHKeySchema,
	}
}
