package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type attachmentModel struct {
	ID       types.String                         `tfsdk:"id"`
	ItemID   types.String                         `tfsdk:"item_id"`
	File     schema_definition.HashedFileValue    `tfsdk:"file"`
	Content  schema_definition.HashedContentValue `tfsdk:"content"`
	FileName types.String                         `tfsdk:"file_name"`
	Size     types.String                         `tfsdk:"size"`
	SizeName types.String                         `tfsdk:"size_name"`
	URL      types.String                         `tfsdk:"url"`
}

var (
	_ resource.Resource                = &attachmentResource{}
	_ resource.ResourceWithConfigure   = &attachmentResource{}
	_ resource.ResourceWithImportState = &attachmentResource{}
)

type attachmentResource struct {
	clients *ProviderClients
}

func NewAttachmentResource() resource.Resource {
	return &attachmentResource{}
}

func (r *attachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attachment"
}

func (r *attachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema_definition.AttachmentResourceSchema()
}

func (r *attachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	r.clients = clients
}

func (r *attachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// State keeps the planned path/content. Replacement is decided at plan
	// time by the digest modifiers.
	var plan attachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	itemId := plan.ItemID.ValueString()

	var obj *models.Attachment
	var err error
	switch {
	case plan.File.ValueString() != "":
		obj, err = bwClient.CreateAttachmentFromFile(ctx, itemId, plan.File.ValueString())
	case plan.Content.ValueString() != "" && plan.FileName.ValueString() != "":
		obj, err = bwClient.CreateAttachmentFromContent(ctx, itemId, plan.FileName.ValueString(), []byte(plan.Content.ValueString()))
	default:
		err = errors.New("either file or content&file_name should be specified")
	}
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	attr := transformation.NewMapData(nil)
	if err = transformation.AttachmentObjectToSchema(ctx, *obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	values := attr.Values()
	plan.ID = types.StringValue(attr.Id())
	plan.FileName = mapStr(values[schema_definition.AttributeAttachmentFileName])
	plan.Size = mapStr(values[schema_definition.AttributeAttachmentSize])
	plan.SizeName = mapStr(values[schema_definition.AttributeAttachmentSizeName])
	plan.URL = mapStr(values[schema_definition.AttributeAttachmentURL])

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *attachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state attachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	itemId := state.ItemID.ValueString()

	// If the item is not found, we can't consider the attachment as simply
	// deleted, because we won't have an item to attach it to. So we surface the
	// error instead of removing the resource from state.
	obj, err := bwClient.GetItem(ctx, models.Item{ID: itemId, Object: models.ObjectTypeItem})
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	for _, attachment := range obj.Attachments {
		if attachment.ID == state.ID.ValueString() {
			attr := transformation.NewMapData(nil)
			if err = transformation.AttachmentObjectToSchema(ctx, attachment, attr); err != nil {
				addErr(&resp.Diagnostics, err)
				return
			}
			values := attr.Values()
			state.ID = types.StringValue(attr.Id())
			state.FileName = mapStr(values[schema_definition.AttributeAttachmentFileName])
			state.Size = mapStr(values[schema_definition.AttributeAttachmentSize])
			state.SizeName = mapStr(values[schema_definition.AttributeAttachmentSizeName])
			state.URL = mapStr(values[schema_definition.AttributeAttachmentURL])
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}

	// The item exists but the attachment is gone: consider it deleted.
	resp.State.RemoveResource(ctx)
}

func (r *attachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Digests unchanged (e.g. same file contents under a new path): no API
	// call, just persist the planned path/content. Content changes force
	// replacement via plan modifiers, so this path is state-only.
	var plan, state attachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.ID.IsUnknown() {
		plan.ID = state.ID
	}
	if plan.FileName.IsUnknown() {
		plan.FileName = state.FileName
	}
	if plan.Size.IsUnknown() {
		plan.Size = state.Size
	}
	if plan.SizeName.IsUnknown() {
		plan.SizeName = state.SizeName
	}
	if plan.URL.IsUnknown() {
		plan.URL = state.URL
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *attachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state attachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	if err := bwClient.DeleteAttachment(ctx, state.ItemID.ValueString(), state.ID.ValueString()); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}
}

func (r *attachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	split := strings.Split(req.ID, "/")
	if len(split) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID should be in the format <attachment_id>/<item_id>: '%s'", req.ID),
		)
		return
	}
	// Format is <attachment_id>/<item_id> (what the SDKv2 importer implemented;
	// older docs had the sides reversed).
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(schema_definition.AttributeID), split[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(schema_definition.AttributeAttachmentItemID), split[1])...)
}
