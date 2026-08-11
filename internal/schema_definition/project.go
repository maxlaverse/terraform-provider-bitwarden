package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ProjectResourceSchema() rsschema.Schema {
	return rsschema.Schema{
		MarkdownDescription: "Manages a Project.",
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
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func ProjectDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Use this data source to get information on an existing project.",
		Attributes: map[string]dsschema.Attribute{
			AttributeID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionIdentifier,
				Optional:            true,
			},
			AttributeName: dsschema.StringAttribute{
				MarkdownDescription: DescriptionName,
				Computed:            true,
			},
			AttributeOrganizationID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionOrganizationID,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}
