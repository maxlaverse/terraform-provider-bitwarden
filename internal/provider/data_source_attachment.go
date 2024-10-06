package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the content on an existing item's attachment.",
		ReadContext: withPasswordManager(resourceReadDataSourceAttachment),
		Schema: map[string]*schema.Schema{
			attributeID: {
				Description: descriptionIdentifier,
				Type:        schema.TypeString,
				Required:    true,
			},
			attributeAttachmentItemID: {
				Description: descriptionItemIdentifier,
				Type:        schema.TypeString,
				Required:    true,
			},
			attributeAttachmentContent: {
				Description: descriptionItemAttachmentContent,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
