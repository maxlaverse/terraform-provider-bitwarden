package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceItemLogin() *schema.Resource {
	dataSourceItemLoginSchema := schema_definition.BaseSchema(schema_definition.DataSource)
	for k, v := range schema_definition.LoginSchema(schema_definition.DataSource) {
		dataSourceItemLoginSchema[k] = v
	}

	return &schema.Resource{
		Description: "Use this data source to get information on an existing login item.",
		ReadContext: withPasswordManager(opItemRead(models.ObjectTypeItem, models.ItemTypeLogin)),
		Schema:      dataSourceItemLoginSchema,
	}
}
