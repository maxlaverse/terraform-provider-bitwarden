package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func secretSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		attributeID: {
			Description: descriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    true,
		},
		attributeKey: {
			Description: descriptionName,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		attributeValue: {
			Description: descriptionValue,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		attributeNote: {
			Description: descriptionNote,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		attributeOrganizationID: {
			Description: descriptionOrganizationID,
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		attributeProjectID: {
			Description: descriptionProjectID,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
	}
}
