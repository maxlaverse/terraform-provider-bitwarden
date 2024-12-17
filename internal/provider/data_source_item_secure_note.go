package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := schema_definition.BaseSchema(schema_definition.DataSource)

	return &schema.Resource{
		Description: "Use this data source to get information on an existing secure note item.",
		ReadContext: withPasswordManager(opItemRead(models.ObjectTypeItem, models.ItemTypeSecureNote)),
		Schema:      dataSourceItemSecureNoteSchema,
	}
}
