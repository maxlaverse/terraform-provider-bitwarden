package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceOrganization() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing organization.",
		ReadContext: withPasswordManager(opObjectRead(models.ObjectTypeOrganization)),
		Schema:      schema_definition.OrganizationSchema(),
	}
}
