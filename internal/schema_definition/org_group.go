package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func OrgGroupSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    true,
		},
		AttributeOrganizationID: {
			Description: DescriptionOrganizationID,
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    schemaType == Resource,
		},
		AttributeName: {
			Description: DescriptionName,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
			ForceNew:    schemaType == Resource,
		},
		AttributeMember: {
			Description: DescriptionOrgGroupMember,
			Type:        schema.TypeSet,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
	}

	if schemaType == DataSource {
		base[AttributeFilterName] = &schema.Schema{
			Description:  DescriptionFilterName,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{AttributeFilterName, AttributeID},
		}
	}

	return base
}
