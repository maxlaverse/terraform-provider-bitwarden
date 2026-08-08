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

func FolderResourceSchema() rsschema.Schema {
	return rsschema.Schema{
		MarkdownDescription: "Manages a folder.",
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
		},
	}
}

func FolderDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Use this data source to get information on an existing folder.",
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
			AttributeFilterCollectionId: dsschema.StringAttribute{
				MarkdownDescription: DescriptionFilterCollectionID,
				Optional:            true,
			},
			AttributeFilterOrganizationID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionFilterOrganizationID,
				Optional:            true,
			},
			AttributeFilterSearch: dsschema.StringAttribute{
				MarkdownDescription: DescriptionFilterSearch,
				Optional:            true,
				Validators: []validator.String{
					fwstringvalidator.AtLeastOneOf(path.MatchRoot(AttributeFilterSearch), path.MatchRoot(AttributeID)),
				},
			},
		},
	}
}
