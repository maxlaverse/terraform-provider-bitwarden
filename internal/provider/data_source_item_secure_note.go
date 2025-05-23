package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceItemSecureNote() *schema.Resource {
	itemSecureNoteSchema := schema_definition.ItemBaseSchema(schema_definition.DataSource)
	for k, v := range schema_definition.SecureNoteSchema(schema_definition.DataSource) {
		itemSecureNoteSchema[k] = v
	}

	return &schema.Resource{
		Description: "Use this data source to get information on an existing secure note item.",
		ReadContext: withPasswordManager(opItemRead(models.ItemTypeSecureNote)),
		Schema:      itemSecureNoteSchema,
	}
}
