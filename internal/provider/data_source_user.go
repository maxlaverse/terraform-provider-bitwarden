package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on a Vaultwarden user account.",
		ReadContext: withPasswordManager(opUserRead),
		Schema:      schema_definition.UserSchema(schema_definition.DataSource),
	}
}
