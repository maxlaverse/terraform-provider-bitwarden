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
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

type projectResource struct {
	clients *ProviderClients
}

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema_definition.ProjectResourceSchema()
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	r.clients = clients
}

func (r *projectResource) projectAttrFromModel(model projectResourceModel) *transformation.MapData {
	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeName:           model.Name.ValueString(),
		schema_definition.AttributeOrganizationID: model.OrganizationID.ValueString(),
	})
	attr.SetId(model.ID.ValueString())
	return attr
}

func projectModelFromData(attr *transformation.MapData) projectResourceModel {
	return projectResourceModel{
		ID:             types.StringValue(attr.Id()),
		Name:           mapStr(attr.Values()[schema_definition.AttributeName]),
		OrganizationID: mapStr(attr.Values()[schema_definition.AttributeOrganizationID]),
	}
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.projectAttrFromModel(plan)
	obj, err := bwsClient.CreateProject(ctx, transformation.ProjectSchemaToObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.ProjectObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectModelFromData(attr))...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.projectAttrFromModel(state)
	obj, err := bwsClient.GetProject(ctx, transformation.ProjectSchemaToObject(ctx, attr))
	if err != nil {
		if errors.Is(err, models.ErrObjectNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.ProjectObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectModelFromData(attr))...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.projectAttrFromModel(plan)
	obj, err := bwsClient.EditProject(ctx, transformation.ProjectSchemaToObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.ProjectObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectModelFromData(attr))...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.projectAttrFromModel(state)
	if err := bwsClient.DeleteProject(ctx, transformation.ProjectSchemaToObject(ctx, attr)); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(schema_definition.AttributeID), req, resp)
}
