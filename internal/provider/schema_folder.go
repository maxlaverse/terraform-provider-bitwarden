package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func folderSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		attributeID: {
			Description: descriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    true,
		},
		attributeName: {
			Description: descriptionName,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		attributeObject: {
			Description: descriptionInternal,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	if schemaType == DataSource {
		base[attributeFilterCollectionId] = &schema.Schema{
			Description: descriptionFilterCollectionID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[attributeFilterOrganizationID] = &schema.Schema{
			Description: descriptionFilterOrganizationID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[attributeFilterSearch] = &schema.Schema{
			Description:  descriptionFilterSearch,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{attributeFilterSearch, attributeID},
		}
	}

	return base
}
