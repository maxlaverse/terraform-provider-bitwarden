package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceOrgCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing organization collection.",
		ReadContext: withPasswordManager(opOrganizationCollectionRead),
		Schema:      schema_definition.OrgCollectionSchema(schema_definition.DataSource),
	}
}
