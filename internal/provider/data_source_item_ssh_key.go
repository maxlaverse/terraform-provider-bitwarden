package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceItemSSHKey() *schema.Resource {
	dataSourceItemSSHKeySchema := schema_definition.ItemBaseSchema(schema_definition.DataSource)
	for k, v := range schema_definition.SSHKeySchema(schema_definition.DataSource) {
		dataSourceItemSSHKeySchema[k] = v
	}

	return &schema.Resource{
		Description: "Use this data source to get information on an existing SSH key item.",
		ReadContext: withPasswordManager(opItemRead(models.ItemTypeSSHKey)),
		Schema:      dataSourceItemSSHKeySchema,
	}
}
