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
	_ resource.Resource                = &secretResource{}
	_ resource.ResourceWithConfigure   = &secretResource{}
	_ resource.ResourceWithImportState = &secretResource{}
)

type secretResource struct {
	clients *ProviderClients
}

func NewSecretResource() resource.Resource {
	return &secretResource{}
}

type secretResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	Note           types.String `tfsdk:"note"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectID      types.String `tfsdk:"project_id"`
}

func (r *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema_definition.SecretResourceSchema()
}

func (r *secretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	r.clients = clients
}

func (r *secretResource) secretAttrFromModel(model secretResourceModel) *transformation.MapData {
	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeKey:            model.Key.ValueString(),
		schema_definition.AttributeValue:          model.Value.ValueString(),
		schema_definition.AttributeNote:           model.Note.ValueString(),
		schema_definition.AttributeOrganizationID: model.OrganizationID.ValueString(),
		schema_definition.AttributeProjectID:      model.ProjectID.ValueString(),
	})
	attr.SetId(model.ID.ValueString())
	return attr
}

func secretModelFromData(attr *transformation.MapData) secretResourceModel {
	vals := attr.Values()
	return secretResourceModel{
		ID:             types.StringValue(attr.Id()),
		Key:            mapStr(vals[schema_definition.AttributeKey]),
		Value:          mapStr(vals[schema_definition.AttributeValue]),
		Note:           mapStr(vals[schema_definition.AttributeNote]),
		OrganizationID: mapStr(vals[schema_definition.AttributeOrganizationID]),
		ProjectID:      mapStr(vals[schema_definition.AttributeProjectID]),
	}
}

func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.secretAttrFromModel(plan)
	obj, err := bwsClient.CreateSecret(ctx, transformation.SecretSchemaToObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.SecretObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, secretModelFromData(attr))...)
}

func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.secretAttrFromModel(state)
	obj, err := bwsClient.GetSecret(ctx, transformation.SecretSchemaToObject(ctx, attr))
	if err != nil {
		if errors.Is(err, models.ErrObjectNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.SecretObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, secretModelFromData(attr))...)
}

func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan secretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.secretAttrFromModel(plan)
	obj, err := bwsClient.EditSecret(ctx, transformation.SecretSchemaToObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.SecretObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, secretModelFromData(attr))...)
}

func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.secretAttrFromModel(state)
	if err := bwsClient.DeleteSecret(ctx, transformation.SecretSchemaToObject(ctx, attr)); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}
}

func (r *secretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(schema_definition.AttributeID), req, resp)
}
