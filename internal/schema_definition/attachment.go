package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func AttachmentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachmentFileName: {
			Description: DescriptionItemAttachmentFileName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachmentSize: {
			Description: DescriptionItemAttachmentSize,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachmentSizeName: {
			Description: DescriptionItemAttachmentSizeName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachmentURL: {
			Description: DescriptionItemAttachmentURL,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
