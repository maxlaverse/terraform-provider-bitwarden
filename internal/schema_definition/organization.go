package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func OrganizationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Optional:    true,
		},
		AttributeName: {
			Description: DescriptionName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeFilterSearch: {
			Description:  DescriptionFilterSearch,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{AttributeFilterSearch, AttributeID},
		},
	}
}
