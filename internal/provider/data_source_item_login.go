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
		Description: "Use this data source to get information on an existing login item.",
		ReadContext: readDataSourceItem(bw.ObjectTypeItem, bw.ItemTypeLogin),
		Schema:      dataSourceItemLoginSchema,
	}
}
