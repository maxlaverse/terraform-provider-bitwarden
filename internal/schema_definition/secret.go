package schema_definition

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func SecretSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	baseSchema := map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    true,
		},
		AttributeKey: {
			Description: DescriptionName,
			Type:        schema.TypeString,
			Optional:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		AttributeValue: {
			Description: DescriptionValue,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		AttributeNote: {
			Description: DescriptionNote,
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
		AttributeProjectID: {
			Description: DescriptionProjectID,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
	}

	if schemaType == DataSource {
		baseSchema[AttributeID].AtLeastOneOf = []string{AttributeID, AttributeKey}
		baseSchema[AttributeID].ConflictsWith = []string{AttributeKey}
		baseSchema[AttributeKey].AtLeastOneOf = []string{AttributeID, AttributeKey}
		baseSchema[AttributeKey].ConflictsWith = []string{AttributeID}
	}

	return baseSchema
}
