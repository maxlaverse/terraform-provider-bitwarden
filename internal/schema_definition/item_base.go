package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type schemaTypeEnum int

const (
	DataSource schemaTypeEnum = 0
	Resource   schemaTypeEnum = 1
)

func ItemBaseSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {

	base := map[string]*schema.Schema{
		/*
		* Attributes that can be required
		 */
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

		/*
		* Most common attributes
		 */
		AttributeCollectionIDs: {
			Description: DescriptionCollectionIDs,
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
		AttributeFavorite: {
			Description: DescriptionFavorite,
			Type:        schema.TypeBool,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
		AttributeField: {
			Description: DescriptionField,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					AttributeFieldName: {
						Description: DescriptionFieldName,
						Type:        schema.TypeString,
						Required:    true,
					},
					AttributeFieldText: {
						Description: DescriptionFieldText,
						Type:        schema.TypeString,
						Optional:    true,
					},
					AttributeFieldBoolean: {
						Description: DescriptionFieldBoolean,
						Type:        schema.TypeBool,
						Optional:    true,
					},
					AttributeFieldHidden: {
						Description: DescriptionFieldHidden,
						Type:        schema.TypeString,
						Optional:    true,
					},
					AttributeFieldLinked: {
						Description: DescriptionFieldLinked,
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
			Computed:  schemaType == DataSource,
			Optional:  schemaType == Resource,
			Sensitive: true,
		},
		AttributeFolderID: {
			Description: DescriptionFolderID,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},

		AttributeNotes: {
			Description: DescriptionNotes,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		AttributeOrganizationID: {
			Description: DescriptionOrganizationID,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
		AttributeReprompt: {
			Description: DescriptionReprompt,
			Type:        schema.TypeBool,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},

		/*
		* Attributes that are always computed
		 */
		AttributeCreationDate: {
			Description: DescriptionCreationDate,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeDeletedDate: {
			Description: DescriptionDeletedDate,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeRevisionDate: {
			Description: DescriptionRevisionDate,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachments: {
			Description: DescriptionAttachments,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: AttachmentSchema(),
			},
			Computed: true,
		},
	}

	if schemaType == DataSource {
		base[AttributeFilterCollectionId] = &schema.Schema{
			Description: DescriptionFilterCollectionID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[AttributeFilterFolderID] = &schema.Schema{
			Description: DescriptionFilterFolderID,
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

func AttachmentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachmentFileName: {
			Description: DescriptionItemAttachmentFileName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachmentSize: {
			Description: DescriptionItemAttachmentSize,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachmentSizeName: {
			Description: DescriptionItemAttachmentSizeName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeAttachmentURL: {
			Description: DescriptionItemAttachmentURL,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
