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

func AttachmentResourceSchema() rsschema.Schema {
	return rsschema.Schema{
		MarkdownDescription: "Manages an item attachment.",
		Attributes: map[string]rsschema.Attribute{
			AttributeID: rsschema.StringAttribute{
				MarkdownDescription: DescriptionIdentifier,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			AttributeAttachmentItemID: rsschema.StringAttribute{
				MarkdownDescription: DescriptionItemIdentifier,
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			AttributeAttachmentFile: rsschema.StringAttribute{
				MarkdownDescription: DescriptionItemAttachmentFile,
				Optional:            true,
				// Digest equality: same bytes under a different path do not
				// replace. Leftover SDKv2 SHA1 state is accepted until rewritten.
				CustomType:    HashedFileType{},
				PlanModifiers: AttachmentFilePlanModifiers(),
				Validators: []validator.String{
					fwstringvalidator.ConflictsWith(path.MatchRoot(AttributeAttachmentContent)),
					fwstringvalidator.AtLeastOneOf(path.MatchRoot(AttributeAttachmentFile), path.MatchRoot(AttributeAttachmentContent)),
				},
			},
			AttributeAttachmentContent: rsschema.StringAttribute{
				MarkdownDescription: DescriptionItemAttachmentContent,
				Optional:            true,
				// Same digest rules as `file`, for the raw body.
				CustomType:    HashedContentType{},
				PlanModifiers: AttachmentContentPlanModifiers(),
				Validators: []validator.String{
					fwstringvalidator.ConflictsWith(path.MatchRoot(AttributeAttachmentFile)),
					fwstringvalidator.AlsoRequires(path.MatchRoot(AttributeAttachmentFileName)),
				},
			},
			AttributeAttachmentFileName: rsschema.StringAttribute{
				MarkdownDescription: DescriptionItemAttachmentFileName,
				Optional:            true,
				Computed:            true,
				// When `file` is set, file_name is computed from the upload
				// response (SDKv2 ComputedWhen). RequiresReplace keeps the
				// ForceNew behaviour of the previous schema.
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
				Validators: []validator.String{
					fwstringvalidator.ConflictsWith(path.MatchRoot(AttributeAttachmentFile)),
				},
			},
			AttributeAttachmentSize: rsschema.StringAttribute{
				MarkdownDescription: DescriptionItemAttachmentSize,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			AttributeAttachmentSizeName: rsschema.StringAttribute{
				MarkdownDescription: DescriptionItemAttachmentSizeName,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			AttributeAttachmentURL: rsschema.StringAttribute{
				MarkdownDescription: DescriptionItemAttachmentURL,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func AttachmentDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Use this data source to get the content on an existing item's attachment.",
		Attributes: map[string]dsschema.Attribute{
			AttributeID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionIdentifier,
				Required:            true,
			},
			AttributeAttachmentItemID: dsschema.StringAttribute{
				MarkdownDescription: DescriptionItemIdentifier,
				Required:            true,
			},
			AttributeAttachmentContent: dsschema.StringAttribute{
				MarkdownDescription: DescriptionItemAttachmentContent,
				Computed:            true,
			},
		},
	}
}
