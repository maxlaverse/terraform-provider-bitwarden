package schema_definition

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ProjectSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		AttributeOrganizationID: {
			Description: DescriptionOrganizationID,
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
	}
}
