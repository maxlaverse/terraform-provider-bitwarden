package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func SecretEphemeralResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Reads an existing Secrets Manager secret without persisting its data in plan or state files. " +
			"Requires Terraform 1.10+ or OpenTofu 1.11+. " +
			"Use the result in an ephemeral context, such as a provider configuration or a write-only resource argument. " +
			"Reading a secret does not create, rotate, or delete it.",

		Attributes: map[string]schema.Attribute{
			AttributeID: schema.StringAttribute{
				MarkdownDescription: "Identifier of the secret. Specify exactly one of `id` or `key`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(path.MatchRoot(AttributeKey)),
				},
			},
			AttributeKey: schema.StringAttribute{
				MarkdownDescription: "Name of the secret. Must uniquely match a secret accessible to the access token. Specify exactly one of `id` or `key`.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			AttributeValue: schema.StringAttribute{
				MarkdownDescription: DescriptionValue,
				Computed:            true,
				Sensitive:           true,
			},
			AttributeNote: schema.StringAttribute{
				MarkdownDescription: DescriptionNote,
				Computed:            true,
				Sensitive:           true,
			},
			AttributeOrganizationID: schema.StringAttribute{
				MarkdownDescription: DescriptionOrganizationID,
				Computed:            true,
			},
			AttributeProjectID: schema.StringAttribute{
				MarkdownDescription: DescriptionProjectID,
				Computed:            true,
			},
		},
	}
}
