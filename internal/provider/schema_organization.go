package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func organizationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		attributeID: {
			Description: descriptionIdentifier,
			Type:        schema.TypeString,
			Optional:    true,
		},
		attributeName: {
			Description: descriptionName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeObject: {
			Description: descriptionInternal,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeFilterSearch: {
			Description:  descriptionFilterSearch,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{attributeFilterSearch, attributeID},
		},
	}
}
