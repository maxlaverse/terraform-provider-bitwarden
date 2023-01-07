package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type schemaTypeEnum int

const (
	DataSource schemaTypeEnum = 0
	Resource   schemaTypeEnum = 1
)

func loginSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		attributeLoginPassword: {
			Description: descriptionLoginPassword,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		attributeLoginUsername: {
			Description: descriptionLoginUsername,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		attributeLoginTotp: {
			Description: descriptionLoginTotp,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		attributeLoginURIs: {
			Description: descriptionLoginUri,
			Type:        schema.TypeList,
			Elem:        uriElem(),
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   false,
		},
	}
}

func baseSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {

	return map[string]*schema.Schema{
		/*
		* Attributes that can be required
		 */
		attributeID: {
			Description: descriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Required:    schemaType == DataSource,
		},
		attributeName: {
			Description: descriptionName,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},

		/*
		* Most common attributes
		 */
		attributeCollectionIDs: {
			Description: descriptionCollectionIDs,
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
		attributeFavorite: {
			Description: descriptionFavorite,
			Type:        schema.TypeBool,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
		attributeField: {
			Description: descriptionField,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					attributeFieldName: {
						Description: descriptionFieldName,
						Type:        schema.TypeString,
						Required:    true,
					},
					attributeFieldText: {
						Description: descriptionFieldText,
						Type:        schema.TypeString,
						Optional:    true,
					},
					attributeFieldBoolean: {
						Description: descriptionFieldBoolean,
						Type:        schema.TypeBool,
						Optional:    true,
					},
					attributeFieldHidden: {
						Description: descriptionFieldHidden,
						Type:        schema.TypeString,
						Optional:    true,
					},
					attributeFieldLinked: {
						Description: descriptionFieldLinked,
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
			Computed:  schemaType == DataSource,
			Optional:  schemaType == Resource,
			Sensitive: true,
		},
		attributeFolderID: {
			Description: descriptionFolderID,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},

		attributeNotes: {
			Description: descriptionNotes,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		attributeOrganizationID: {
			Description: descriptionOrganizationID,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
		attributeReprompt: {
			Description: descriptionReprompt,
			Type:        schema.TypeBool,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},

		/*
		* Attributes that are always computed
		 */
		attributeObject: {
			Description: descriptionInternal,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeRevisionDate: {
			Description: descriptionRevisionDate,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeType: {
			Description: descriptionInternal,
			Type:        schema.TypeInt,
			Computed:    true,
		},
		attributeAttachments: {
			Description: descriptionAttachments,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: attachmentSchema(),
			},
			Computed: true,
		},
	}
}

func uriElem() *schema.Resource {
	validMatchStr := []string{"default", "base_domain", "host", "start_with", "exact", "regexp", "never"}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			attributeLoginURIsMatch: {
				Description:      descriptionLoginUriMatch,
				Type:             schema.TypeString,
				Default:          validMatchStr[0],
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(validMatchStr, false)),
				Optional:         true,
			},
			attributeLoginURIsValue: {
				Description: descriptionLoginUriValue,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func attachmentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		attributeID: {
			Description: descriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeAttachmentFileName: {
			Description: descriptionItemAttachmentFileName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeAttachmentSize: {
			Description: descriptionItemAttachmentSize,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeAttachmentSizeName: {
			Description: descriptionItemAttachmentSizeName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeAttachmentURL: {
			Description: descriptionItemAttachmentURL,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
