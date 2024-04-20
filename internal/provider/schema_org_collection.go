package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func orgCollectionSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
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
		attributeOrganizationID: {
			Description: descriptionOrganizationID,
			Type:        schema.TypeString,
			Required:    true,
		},
	}

	if schemaType == DataSource {
		base[attributeFilterSearch] = &schema.Schema{
			Description:  descriptionFilterSearch,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{attributeFilterSearch, attributeID},
		}
	}

	return base
}
