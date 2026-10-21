package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SecretsEphemeralResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Reads selected Secrets Manager secrets without persisting their data in plan or state files. " +
			"Requires Terraform 1.10+ or OpenTofu 1.11+. " +
			"Use the result in an ephemeral context, such as a provider configuration or a write-only resource argument. " +
			"The CLI client lists all accessible secrets before selecting the requested IDs.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of secret IDs to read. IDs are case-insensitive. All requested secrets must exist and be accessible to the access token.",
				ElementType:         types.StringType,
				Required:            true,
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"secrets": schema.MapNestedAttribute{
				MarkdownDescription: "Secrets keyed by canonical lowercase secret IDs.",
				Computed:            true,
				Sensitive:           true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":   schema.StringAttribute{MarkdownDescription: DescriptionName, Computed: true},
						"value": schema.StringAttribute{MarkdownDescription: DescriptionValue, Computed: true, Sensitive: true},
						"note":  schema.StringAttribute{MarkdownDescription: DescriptionNote, Computed: true, Sensitive: true},
					},
				},
			},
		},
	}
}
