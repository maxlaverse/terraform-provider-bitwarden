package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func dataSourceFolder() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing folder.",
		ReadContext: readDataSourceObject(bw.ObjectTypeFolder),
		Schema:      folderSchema(DataSource),
	}
}
