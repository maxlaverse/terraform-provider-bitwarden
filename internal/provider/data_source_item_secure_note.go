package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func dataSourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(DataSource)

	return &schema.Resource{
		Description: "Use this data source to get (amongst other things) the content of a Bitwarden Secret Note for use in other resources.",
		ReadContext: readDataSource(bw.ObjectTypeItem, bw.ItemTypeSecureNote),
		Schema:      dataSourceItemSecureNoteSchema,
	}
}
