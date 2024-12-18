package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the content on an existing item's attachment.",
		ReadContext: withPasswordManager(opAttachmentRead),
		Schema: map[string]*schema.Schema{
			schema_definition.AttributeID: {
				Description: schema_definition.DescriptionIdentifier,
				Type:        schema.TypeString,
				Required:    true,
			},
			schema_definition.AttributeAttachmentItemID: {
				Description: schema_definition.DescriptionItemIdentifier,
				Type:        schema.TypeString,
				Required:    true,
			},
			schema_definition.AttributeAttachmentContent: {
				Description: schema_definition.DescriptionItemAttachmentContent,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
