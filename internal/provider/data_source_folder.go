package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFolder() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing Folder.",
		ReadContext: readDataSourceFolder(),
		Schema:      folderSchema(DataSource),
	}
}
