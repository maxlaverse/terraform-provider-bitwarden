package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SecureNoteSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeFavorite: {
			Description: DescriptionFavorite,
			Type:        schema.TypeBool,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
		AttributeAttachments: {
			Description: DescriptionAttachments,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: AttachmentSchema(),
			},
			Computed: true,
		},
	}

	return base
}
