package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func FolderSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    true,
		},
		AttributeName: {
			Description: DescriptionName,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		AttributeObject: {
			Description: DescriptionInternal,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	if schemaType == DataSource {
		base[AttributeFilterCollectionId] = &schema.Schema{
			Description: DescriptionFilterCollectionID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[AttributeFilterOrganizationID] = &schema.Schema{
			Description: DescriptionFilterOrganizationID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[AttributeFilterSearch] = &schema.Schema{
			Description:  DescriptionFilterSearch,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{AttributeFilterSearch, AttributeID},
		}
	}

	return base
}
