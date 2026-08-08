package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	fwstringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

func OrganizationDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Use this data source to get information on an existing organization.",
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
