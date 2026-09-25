package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

var (
	_ resource.Resource                = &orgCollectionResource{}
	_ resource.ResourceWithConfigure   = &orgCollectionResource{}
	_ resource.ResourceWithImportState = &orgCollectionResource{}
	_ resource.ResourceWithModifyPlan  = &orgCollectionResource{}
)

type orgCollectionResource struct {
	clients *ProviderClients
}

func NewOrgCollectionResource() resource.Resource {
	return &orgCollectionResource{}
}

type orgCollectionResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Member         types.Set    `tfsdk:"member"`
	MemberGroup    types.Set    `tfsdk:"member_group"`
}

func (r *orgCollectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_collection"
}

func (r *orgCollectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema_definition.OrgCollectionResourceSchema()
}

func (r *orgCollectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	r.clients = clients
}

func (r *orgCollectionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var config orgCollectionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan orgCollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var d diag.Diagnostics
	plan.Member, d = normalizeMembershipSet(ctx, config.Member)
	resp.Diagnostics.Append(d...)
	plan.MemberGroup, d = normalizeMembershipSet(ctx, config.MemberGroup)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *orgCollectionResource) orgCollectionAttrFromModel(model orgCollectionResourceModel) *transformation.MapData {
	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeName:           model.Name.ValueString(),
		schema_definition.AttributeOrganizationID: model.OrganizationID.ValueString(),
		schema_definition.AttributeMember:         membershipSetToData(model.Member),
		schema_definition.AttributeMemberGroup:    membershipSetToData(model.MemberGroup),
	})
	attr.SetId(model.ID.ValueString())
	return attr
}

func orgCollectionModelFromData(attr *transformation.MapData) orgCollectionResourceModel {
	vals := attr.Values()
	return orgCollectionResourceModel{
		ID:             types.StringValue(attr.Id()),
		Name:           mapStr(vals[schema_definition.AttributeName]),
		OrganizationID: mapStr(vals[schema_definition.AttributeOrganizationID]),
		Member:         membershipDataToSet(vals[schema_definition.AttributeMember]),
		MemberGroup:    membershipDataToSet(vals[schema_definition.AttributeMemberGroup]),
	}
}

func (r *orgCollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan orgCollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.orgCollectionAttrFromModel(plan)
	obj, err := bwClient.CreateOrganizationCollection(ctx, transformation.OrganizationCollectionToObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.OrganizationCollectionObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	// Keep planned memberships in state: the API may return equivalent
	// members that don't byte-match the planned set (null vs false, etc.),
	// which Framework treats as an apply inconsistency on SetNestedBlock.
	state := orgCollectionModelFromData(attr)
	state.Member = plan.Member
	state.MemberGroup = plan.MemberGroup
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *orgCollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state orgCollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.orgCollectionAttrFromModel(state)
	obj, err := bwClient.GetOrganizationCollection(ctx, transformation.OrganizationCollectionToObject(ctx, attr))
	if err != nil {
		if errors.Is(err, models.ErrObjectNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.OrganizationCollectionObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, orgCollectionModelFromData(attr))...)
}

func (r *orgCollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan orgCollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.orgCollectionAttrFromModel(plan)
	obj, err := bwClient.EditOrganizationCollection(ctx, transformation.OrganizationCollectionToObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.OrganizationCollectionObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	state := orgCollectionModelFromData(attr)
	state.Member = plan.Member
	state.MemberGroup = plan.MemberGroup
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *orgCollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state orgCollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.orgCollectionAttrFromModel(state)
	if err := bwClient.DeleteOrganizationCollection(ctx, transformation.OrganizationCollectionToObject(ctx, attr)); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}
}

func (r *orgCollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	split := strings.Split(req.ID, "/")
	if len(split) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("invalid ID specified, should be in the format <organization_id>/<collection_id>: '%s'", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(schema_definition.AttributeOrganizationID), split[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(schema_definition.AttributeID), split[1])...)
}
