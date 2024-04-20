package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func dataSourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(DataSource)

	return &schema.Resource{
		Description: "Use this data source to get information on an existing Secure Note.",
		ReadContext: readDataSourceItem(bw.ObjectTypeItem, bw.ItemTypeSecureNote),
		Schema:      dataSourceItemSecureNoteSchema,
	}
}
