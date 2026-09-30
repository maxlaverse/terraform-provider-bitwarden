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

const (
	DescriptionSecretIdentifier = "Identifier of the secret. Specify exactly one of `id` or `key`."
	DescriptionSecretKey        = "Name of the secret. Must uniquely match a secret accessible to the access token. Specify exactly one of `id` or `key`."
)

func SecretResourceSchema() rsschema.Schema {
	return rsschema.Schema{
		MarkdownDescription: "Manages a secret.",
		Attributes: map[string]rsschema.Attribute{
			AttributeID: rsschema.StringAttribute{
				MarkdownDescription: DescriptionIdentifier,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			AttributeKey: rsschema.StringAttribute{
				MarkdownDescription: DescriptionName,
				Required:            true,
			},
			AttributeValue: rsschema.StringAttribute{
				MarkdownDescription: DescriptionValue,
				Required:            true,
			},
			AttributeNote: rsschema.StringAttribute{
				MarkdownDescription: DescriptionNote,
				Required:            true,
			},
			AttributeOrganizationID: rsschema.StringAttribute{
				MarkdownDescription: DescriptionOrganizationID,
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			AttributeProjectID: rsschema.StringAttribute{
				MarkdownDescription: DescriptionProjectID,
				Required:            true,
			},
		},
	}
}

func SecretDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Use this data source to get information on an existing secret.",
		Attributes: map[string]dsschema.Attribute{
			AttributeID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionSecretIdentifier,
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					fwstringvalidator.LengthAtLeast(1),
					fwstringvalidator.ExactlyOneOf(path.MatchRoot(AttributeKey)),
				},
			},
			AttributeKey: dsschema.StringAttribute{
				MarkdownDescription: DescriptionSecretKey,
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{fwstringvalidator.LengthAtLeast(1)},
			},
			AttributeValue: dsschema.StringAttribute{
				MarkdownDescription: DescriptionValue,
				Computed:            true,
			},
			AttributeNote: dsschema.StringAttribute{
				MarkdownDescription: DescriptionNote,
				Computed:            true,
			},
			AttributeOrganizationID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionOrganizationID,
				Optional:            true,
				Computed:            true,
			},
			AttributeProjectID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionProjectID,
				Computed:            true,
			},
		},
	}
}
