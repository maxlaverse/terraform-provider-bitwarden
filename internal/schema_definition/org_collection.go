package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	fwstringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

func OrgCollectionResourceSchema() rsschema.Schema {
	return rsschema.Schema{
		MarkdownDescription: "Manages an organization collection.",
		Attributes: map[string]rsschema.Attribute{
			AttributeID: rsschema.StringAttribute{
				MarkdownDescription: DescriptionIdentifier,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			AttributeName: rsschema.StringAttribute{
				MarkdownDescription: DescriptionName,
				Required:            true,
			},
			AttributeOrganizationID: rsschema.StringAttribute{
				MarkdownDescription: DescriptionOrganizationID,
				Required:            true,
			},
		},
		Blocks: map[string]rsschema.Block{
			AttributeMember:      membershipResourceBlock(DescriptionCollectionMember),
			AttributeMemberGroup: membershipResourceBlock(DescriptionCollectionMemberGroup),
		},
	}
}

func OrgCollectionDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Use this data source to get information on an existing organization collection.",
		Attributes: map[string]dsschema.Attribute{
			AttributeID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionIdentifier,
				Optional:            true,
				Computed:            true,
			},
			AttributeName: dsschema.StringAttribute{
				MarkdownDescription: DescriptionName,
				Computed:            true,
			},
			AttributeOrganizationID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionOrganizationID,
				Required:            true,
			},
			AttributeFilterSearch: dsschema.StringAttribute{
				MarkdownDescription: DescriptionFilterSearch,
				Optional:            true,
				Validators: []validator.String{
					fwstringvalidator.AtLeastOneOf(path.MatchRoot(AttributeFilterSearch), path.MatchRoot(AttributeID)),
				},
			},
			AttributeMember:      membershipDataSourceAttribute(DescriptionCollectionMember),
			AttributeMemberGroup: membershipDataSourceAttribute(DescriptionCollectionMemberGroup),
		},
	}
}

func membershipResourceBlock(description string) rsschema.SetNestedBlock {
	return rsschema.SetNestedBlock{
		MarkdownDescription: description,
		NestedObject: rsschema.NestedBlockObject{
			Attributes: map[string]rsschema.Attribute{
				AttributeID: rsschema.StringAttribute{
					MarkdownDescription: DescriptionCollectionMemberID,
					Required:            true,
				},
				AttributeCollectionMemberReadOnly: rsschema.BoolAttribute{
					MarkdownDescription: DescriptionCollectionMemberReadOnly,
					Optional:            true,
					Computed:            true,
				},
				AttributeCollectionMemberHidePasswords: rsschema.BoolAttribute{
					MarkdownDescription: DescriptionCollectionMemberHidePasswords,
					Optional:            true,
					Computed:            true,
				},
				AttributeCollectionMemberManage: rsschema.BoolAttribute{
					MarkdownDescription: DescriptionCollectionMemberManage,
					Optional:            true,
					Computed:            true,
				},
			},
		},
	}
}

func membershipDataSourceAttribute(description string) dsschema.SetNestedAttribute {
	return dsschema.SetNestedAttribute{
		MarkdownDescription: description,
		Computed:            true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				AttributeID: dsschema.StringAttribute{
					MarkdownDescription: DescriptionCollectionMemberID,
					Computed:            true,
				},
				AttributeCollectionMemberReadOnly: dsschema.BoolAttribute{
					MarkdownDescription: DescriptionCollectionMemberReadOnly,
					Computed:            true,
				},
				AttributeCollectionMemberHidePasswords: dsschema.BoolAttribute{
					MarkdownDescription: DescriptionCollectionMemberHidePasswords,
					Computed:            true,
				},
				AttributeCollectionMemberManage: dsschema.BoolAttribute{
					MarkdownDescription: DescriptionCollectionMemberManage,
					Computed:            true,
				},
			},
		},
	}
}
