package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

var (
	_ resource.Resource                = &folderResource{}
	_ resource.ResourceWithConfigure   = &folderResource{}
	_ resource.ResourceWithImportState = &folderResource{}
)

type folderResource struct {
	clients *ProviderClients
}

func NewFolderResource() resource.Resource {
	return &folderResource{}
}

type folderResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *folderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

func (r *folderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema_definition.FolderResourceSchema()
}

func (r *folderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	r.clients = clients
}

func (r *folderResource) folderAttrFromModel(model folderResourceModel) *transformation.MapData {
	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeName: model.Name.ValueString(),
	})
	attr.SetId(model.ID.ValueString())
	return attr
}

func folderModelFromData(attr *transformation.MapData) folderResourceModel {
	return folderResourceModel{
		ID:   types.StringValue(attr.Id()),
		Name: mapStr(attr.Values()[schema_definition.AttributeName]),
	}
}

func (r *folderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan folderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.folderAttrFromModel(plan)
	obj, err := bwClient.CreateFolder(ctx, transformation.SchemaToFolderObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.FolderObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, folderModelFromData(attr))...)
}

func (r *folderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state folderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.folderAttrFromModel(state)
	obj, err := bwClient.GetFolder(ctx, transformation.SchemaToFolderObject(ctx, attr))
	if err != nil {
		if errors.Is(err, models.ErrObjectNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.FolderObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, folderModelFromData(attr))...)
}

func (r *folderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan folderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.folderAttrFromModel(plan)
	obj, err := bwClient.EditFolder(ctx, transformation.SchemaToFolderObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.FolderObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, folderModelFromData(attr))...)
}

func (r *folderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state folderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.folderAttrFromModel(state)
	if err := bwClient.DeleteFolder(ctx, transformation.SchemaToFolderObject(ctx, attr)); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}
}

func (r *folderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(schema_definition.AttributeID), req, resp)
}
