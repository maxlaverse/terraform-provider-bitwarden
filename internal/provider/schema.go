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
	base := map[string]*schema.Schema{
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

	if schemaType == DataSource {
		base[attributeFilterURL] = &schema.Schema{
			Description: descriptionFilterURL,
			Type:        schema.TypeString,
			Optional:    true,
		}
	}
	return base
}

func baseSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {

	base := map[string]*schema.Schema{
		/*
		* Attributes that can be required
		 */
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
		attributeCreationDate: {
			Description: descriptionCreationDate,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeDeletedDate: {
			Description: descriptionDeletedDate,
			Type:        schema.TypeString,
			Computed:    true,
		},
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

	if schemaType == DataSource {
		base[attributeFilterCollectionId] = &schema.Schema{
			Description: descriptionFilterCollectionID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[attributeFilterFolderID] = &schema.Schema{
			Description: descriptionFilterFolderID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[attributeFilterOrganizationID] = &schema.Schema{
			Description: descriptionFilterOrganizationID,
			Type:        schema.TypeString,
			Optional:    true,
		}

		base[attributeFilterSearch] = &schema.Schema{
			Description:  descriptionFilterSearch,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{attributeFilterSearch, attributeID},
		}
	}
	return base
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
