package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Vaultwarden user account. The created account is in the 'Invited' state until the invitee completes signup.",

		CreateContext: withPasswordManager(opUserCreate),
		ReadContext:   withPasswordManager(opUserReadIgnoreMissing),
		DeleteContext: withPasswordManager(opUserDelete),
		Importer:      resourceImporter(opUserImport),

		Schema: schema_definition.UserSchema(schema_definition.Resource),
	}
}
