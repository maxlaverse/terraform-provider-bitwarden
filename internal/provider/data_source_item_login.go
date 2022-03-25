package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func dataSourceItemLogin() *schema.Resource {
	dataSourceItemLoginSchema := baseSchema(DataSource)
	for k, v := range loginSchema(DataSource) {
		dataSourceItemLoginSchema[k] = v
	}

	return &schema.Resource{
		Description: "Use this data source to get (amongst other things) the username and password of a Bitwarden Login item for use in other resources.",
		ReadContext: readDataSource(bw.ObjectTypeItem, bw.ItemTypeLogin),
		Schema:      dataSourceItemLoginSchema,
	}
}
