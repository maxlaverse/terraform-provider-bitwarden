package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func OrgGroupSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Optional:    true,
		},
		AttributeOrganizationID: {
			Description: DescriptionOrganizationID,
			Type:        schema.TypeString,
			Required:    true,
		},
		AttributeName: {
			Description: DescriptionName,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	if schemaType == DataSource {
		base[AttributeFilterOrganizationID] = &schema.Schema{
			Description: DescriptionFilterOrganizationID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[AttributeFilterName] = &schema.Schema{
			Description:  DescriptionFilterName,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{AttributeFilterName, AttributeID},
		}
	}

	return base
}
