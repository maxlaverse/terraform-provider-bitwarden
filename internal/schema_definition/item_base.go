package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	fwstringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

func fieldResourceBlock() rsschema.ListNestedBlock {
	// Nested attrs are all Sensitive: ListNestedBlock cannot mark the list itself
	// (SDKv2 field list was wholly Sensitive).
	return rsschema.ListNestedBlock{
		Description: DescriptionField,
		NestedObject: rsschema.NestedBlockObject{
			Attributes: map[string]rsschema.Attribute{
				AttributeFieldName: rsschema.StringAttribute{
					Description: DescriptionFieldName,
					Required:    true,
					Sensitive:   true,
				},
				AttributeFieldText: rsschema.StringAttribute{
					Description: DescriptionFieldText,
					Optional:    true,
					Sensitive:   true,
				},
				AttributeFieldBoolean: rsschema.BoolAttribute{
					Description: DescriptionFieldBoolean,
					Optional:    true,
					Sensitive:   true,
				},
				AttributeFieldHidden: rsschema.StringAttribute{
					Description: DescriptionFieldHidden,
					Optional:    true,
					Sensitive:   true,
				},
				AttributeFieldLinked: rsschema.StringAttribute{
					Description: DescriptionFieldLinked,
					Optional:    true,
					Sensitive:   true,
				},
			},
		},
	}
}

func fieldDataSourceAttribute() dsschema.ListNestedAttribute {
	return dsschema.ListNestedAttribute{
		Description: DescriptionField,
		Computed:    true,
		Sensitive:   true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				AttributeFieldName:    dsschema.StringAttribute{Description: DescriptionFieldName, Computed: true},
				AttributeFieldText:    dsschema.StringAttribute{Description: DescriptionFieldText, Computed: true},
				AttributeFieldBoolean: dsschema.BoolAttribute{Description: DescriptionFieldBoolean, Computed: true},
				AttributeFieldHidden:  dsschema.StringAttribute{Description: DescriptionFieldHidden, Computed: true},
				AttributeFieldLinked:  dsschema.StringAttribute{Description: DescriptionFieldLinked, Computed: true},
			},
		},
	}
}

func uriResourceBlock() rsschema.ListNestedBlock {
	return rsschema.ListNestedBlock{
		Description: DescriptionLoginUri,
		NestedObject: rsschema.NestedBlockObject{
			Attributes: map[string]rsschema.Attribute{
				AttributeLoginURIsMatch: rsschema.StringAttribute{
					Description: DescriptionLoginUriMatch,
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(string(URIMatchDefaultStr)),
					Validators: []validator.String{
						fwstringvalidator.OneOf(ValidURIMatchStrings()...),
					},
				},
				AttributeLoginURIsValue: rsschema.StringAttribute{
					Description: DescriptionLoginUriValue,
					Required:    true,
				},
			},
		},
	}
}

func uriDataSourceAttribute() dsschema.ListNestedAttribute {
	return dsschema.ListNestedAttribute{
		Description: DescriptionLoginUri,
		Computed:    true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				AttributeLoginURIsMatch: dsschema.StringAttribute{Description: DescriptionLoginUriMatch, Computed: true},
				AttributeLoginURIsValue: dsschema.StringAttribute{Description: DescriptionLoginUriValue, Computed: true},
			},
		},
	}
}

func attachmentsResourceAttribute() rsschema.ListNestedAttribute {
	return rsschema.ListNestedAttribute{
		Description:   DescriptionAttachments,
		Computed:      true,
		PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
		NestedObject: rsschema.NestedAttributeObject{
			Attributes: map[string]rsschema.Attribute{
				AttributeID:                 rsschema.StringAttribute{Description: DescriptionIdentifier, Computed: true},
				AttributeAttachmentFileName: rsschema.StringAttribute{Description: DescriptionItemAttachmentFileName, Computed: true},
				AttributeAttachmentSize:     rsschema.StringAttribute{Description: DescriptionItemAttachmentSize, Computed: true},
				AttributeAttachmentSizeName: rsschema.StringAttribute{Description: DescriptionItemAttachmentSizeName, Computed: true},
				AttributeAttachmentURL:      rsschema.StringAttribute{Description: DescriptionItemAttachmentURL, Computed: true},
			},
		},
	}
}

func attachmentsDataSourceAttribute() dsschema.ListNestedAttribute {
	return dsschema.ListNestedAttribute{
		Description: DescriptionAttachments,
		Computed:    true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				AttributeID:                 dsschema.StringAttribute{Description: DescriptionIdentifier, Computed: true},
				AttributeAttachmentFileName: dsschema.StringAttribute{Description: DescriptionItemAttachmentFileName, Computed: true},
				AttributeAttachmentSize:     dsschema.StringAttribute{Description: DescriptionItemAttachmentSize, Computed: true},
				AttributeAttachmentSizeName: dsschema.StringAttribute{Description: DescriptionItemAttachmentSizeName, Computed: true},
				AttributeAttachmentURL:      dsschema.StringAttribute{Description: DescriptionItemAttachmentURL, Computed: true},
			},
		},
	}
}

func itemBaseResourceAttributes() map[string]rsschema.Attribute {
	return map[string]rsschema.Attribute{
		AttributeID: rsschema.StringAttribute{
			Description:   DescriptionIdentifier,
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		AttributeName: rsschema.StringAttribute{
			Description: DescriptionName,
			Required:    true,
		},
		AttributeCollectionIDs: rsschema.SetAttribute{
			Description: DescriptionCollectionIDs,
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
		},
		AttributeFolderID: rsschema.StringAttribute{
			Description: DescriptionFolderID,
			Optional:    true,
			Computed:    true,
		},
		AttributeNotes: rsschema.StringAttribute{
			Description: DescriptionNotes,
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
		},
		AttributeOrganizationID: rsschema.StringAttribute{
			Description:   DescriptionOrganizationID,
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		AttributeReprompt: rsschema.BoolAttribute{
			Description: DescriptionReprompt,
			Optional:    true,
			Computed:    true,
		},
		AttributeCreationDate: rsschema.StringAttribute{
			Description: DescriptionCreationDate,
			Computed:    true,
		},
		AttributeDeletedDate: rsschema.StringAttribute{
			Description: DescriptionDeletedDate,
			Computed:    true,
		},
		AttributeRevisionDate: rsschema.StringAttribute{
			Description: DescriptionRevisionDate,
			Computed:    true,
		},
	}
}

func itemBaseDataSourceAttributes() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		AttributeID: dsschema.StringAttribute{
			Description: DescriptionIdentifier,
			Optional:    true,
			Computed:    true,
		},
		AttributeName: dsschema.StringAttribute{
			Description: DescriptionName,
			Computed:    true,
		},
		AttributeCollectionIDs: dsschema.SetAttribute{
			Description: DescriptionCollectionIDs,
			ElementType: types.StringType,
			Computed:    true,
		},
		AttributeFolderID: dsschema.StringAttribute{
			Description: DescriptionFolderID,
			Computed:    true,
		},
		AttributeNotes: dsschema.StringAttribute{
			Description: DescriptionNotes,
			Computed:    true,
			Sensitive:   true,
		},
		AttributeOrganizationID: dsschema.StringAttribute{
			Description: DescriptionOrganizationID,
			Computed:    true,
		},
		AttributeReprompt: dsschema.BoolAttribute{
			Description: DescriptionReprompt,
			Computed:    true,
		},
		AttributeCreationDate: dsschema.StringAttribute{Description: DescriptionCreationDate, Computed: true},
		AttributeDeletedDate:  dsschema.StringAttribute{Description: DescriptionDeletedDate, Computed: true},
		AttributeRevisionDate: dsschema.StringAttribute{Description: DescriptionRevisionDate, Computed: true},
		AttributeField:        fieldDataSourceAttribute(),

		AttributeFilterCollectionId: dsschema.StringAttribute{
			Description: DescriptionFilterCollectionID,
			Optional:    true,
		},
		AttributeFilterFolderID: dsschema.StringAttribute{
			Description: DescriptionFilterFolderID,
			Optional:    true,
		},
		AttributeFilterOrganizationID: dsschema.StringAttribute{
			Description: DescriptionFilterOrganizationID,
			Optional:    true,
		},
		AttributeFilterSearch: dsschema.StringAttribute{
			Description: DescriptionFilterSearch,
			Optional:    true,
			Validators: []validator.String{
				fwstringvalidator.AtLeastOneOf(path.MatchRoot(AttributeFilterSearch), path.MatchRoot(AttributeID)),
			},
		},
	}
}
