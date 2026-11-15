package schema_definition

import (
	epschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	fwstringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

func fieldEphemeralAttribute() epschema.ListNestedAttribute {
	return epschema.ListNestedAttribute{
		Description: DescriptionField,
		Computed:    true,
		Sensitive:   true,
		NestedObject: epschema.NestedAttributeObject{
			Attributes: map[string]epschema.Attribute{
				AttributeFieldName:    epschema.StringAttribute{Description: DescriptionFieldName, Computed: true},
				AttributeFieldText:    epschema.StringAttribute{Description: DescriptionFieldText, Computed: true},
				AttributeFieldBoolean: epschema.BoolAttribute{Description: DescriptionFieldBoolean, Computed: true},
				AttributeFieldHidden:  epschema.StringAttribute{Description: DescriptionFieldHidden, Computed: true},
				AttributeFieldLinked:  epschema.StringAttribute{Description: DescriptionFieldLinked, Computed: true},
			},
		},
	}
}

func uriEphemeralAttribute() epschema.ListNestedAttribute {
	return epschema.ListNestedAttribute{
		Description: DescriptionLoginUri,
		Computed:    true,
		NestedObject: epschema.NestedAttributeObject{
			Attributes: map[string]epschema.Attribute{
				AttributeLoginURIsMatch: epschema.StringAttribute{Description: DescriptionLoginUriMatch, Computed: true},
				AttributeLoginURIsValue: epschema.StringAttribute{Description: DescriptionLoginUriValue, Computed: true},
			},
		},
	}
}

func attachmentsEphemeralAttribute() epschema.ListNestedAttribute {
	return epschema.ListNestedAttribute{
		Description: DescriptionAttachments,
		Computed:    true,
		NestedObject: epschema.NestedAttributeObject{
			Attributes: map[string]epschema.Attribute{
				AttributeID:                 epschema.StringAttribute{Description: DescriptionIdentifier, Computed: true},
				AttributeAttachmentFileName: epschema.StringAttribute{Description: DescriptionItemAttachmentFileName, Computed: true},
				AttributeAttachmentSize:     epschema.StringAttribute{Description: DescriptionItemAttachmentSize, Computed: true},
				AttributeAttachmentSizeName: epschema.StringAttribute{Description: DescriptionItemAttachmentSizeName, Computed: true},
				AttributeAttachmentURL:      epschema.StringAttribute{Description: DescriptionItemAttachmentURL, Computed: true},
			},
		},
	}
}

func itemBaseEphemeralAttributes() map[string]epschema.Attribute {
	return map[string]epschema.Attribute{
		AttributeID: epschema.StringAttribute{
			Description: DescriptionIdentifier,
			Optional:    true,
			Computed:    true,
		},
		AttributeName: epschema.StringAttribute{
			Description: DescriptionName,
			Computed:    true,
		},
		AttributeCollectionIDs: epschema.SetAttribute{
			Description: DescriptionCollectionIDs,
			ElementType: types.StringType,
			Computed:    true,
		},
		AttributeFolderID: epschema.StringAttribute{
			Description: DescriptionFolderID,
			Computed:    true,
		},
		AttributeNotes: epschema.StringAttribute{
			Description: DescriptionNotes,
			Computed:    true,
			Sensitive:   true,
		},
		AttributeOrganizationID: epschema.StringAttribute{
			Description: DescriptionOrganizationID,
			Computed:    true,
		},
		AttributeReprompt: epschema.BoolAttribute{
			Description: DescriptionReprompt,
			Computed:    true,
		},
		AttributeCreationDate: epschema.StringAttribute{Description: DescriptionCreationDate, Computed: true},
		AttributeDeletedDate:  epschema.StringAttribute{Description: DescriptionDeletedDate, Computed: true},
		AttributeRevisionDate: epschema.StringAttribute{Description: DescriptionRevisionDate, Computed: true},
		AttributeField:        fieldEphemeralAttribute(),

		AttributeFilterCollectionId: epschema.StringAttribute{
			Description: DescriptionFilterCollectionID,
			Optional:    true,
		},
		AttributeFilterFolderID: epschema.StringAttribute{
			Description: DescriptionFilterFolderID,
			Optional:    true,
		},
		AttributeFilterOrganizationID: epschema.StringAttribute{
			Description: DescriptionFilterOrganizationID,
			Optional:    true,
		},
		AttributeFilterSearch: epschema.StringAttribute{
			Description: DescriptionFilterSearch,
			Optional:    true,
			Validators: []validator.String{
				fwstringvalidator.AtLeastOneOf(path.MatchRoot(AttributeFilterSearch), path.MatchRoot(AttributeID)),
			},
		},
	}
}
