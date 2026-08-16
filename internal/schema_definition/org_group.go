package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	fwstringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

func OrgGroupDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Use this data source to get information on an existing organization group.",
		Attributes: map[string]dsschema.Attribute{
			AttributeID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionIdentifier,
				Optional:            true,
				Computed:            true,
			},
			AttributeOrganizationID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionOrganizationID,
				Required:            true,
			},
			AttributeName: dsschema.StringAttribute{
				MarkdownDescription: DescriptionName,
				Computed:            true,
			},
			AttributeFilterName: dsschema.StringAttribute{
				MarkdownDescription: DescriptionFilterName,
				Optional:            true,
				Validators: []validator.String{
					fwstringvalidator.AtLeastOneOf(path.MatchRoot(AttributeFilterName), path.MatchRoot(AttributeID)),
				},
			},
		},
	}
}
