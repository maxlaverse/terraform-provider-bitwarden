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
		MarkdownDescription: "Reads selected Secrets Manager secrets in one batch without persisting their values in plan or state files. " +
			"Requires Terraform 1.10+ or OpenTofu 1.11+. The embedded client fetches only the requested IDs in one API call; " +
			"the CLI client runs bws secret list once and filters all accessible secrets locally. " +
			"Secrets are read again in each phase; this resource does not cache values between plan and apply or skip unchanged secrets.",
		Attributes: map[string]schema.Attribute{
			"ids": schema.SetAttribute{
				MarkdownDescription: "Set of secret IDs to read. Only these secrets are returned. IDs are case-insensitive and each secret is fetched once. An empty set performs no secret retrieval request. All requested secrets must exist and be accessible to the access token.",
				ElementType:         types.StringType,
				Required:            true,
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"values": schema.MapAttribute{
				MarkdownDescription: "Secret values keyed by canonical lowercase secret IDs. Use in an ephemeral context, such as a write-only resource argument.",
				ElementType:         types.StringType,
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}
