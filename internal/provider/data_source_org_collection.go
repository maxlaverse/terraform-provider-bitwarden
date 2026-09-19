package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

var (
	_ datasource.DataSource              = &orgCollectionDataSource{}
	_ datasource.DataSourceWithConfigure = &orgCollectionDataSource{}
)

type orgCollectionDataSource struct {
	clients *ProviderClients
}

func NewOrgCollectionDataSource() datasource.DataSource {
	return &orgCollectionDataSource{}
}

type orgCollectionDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Search         types.String `tfsdk:"search"`
	Member         types.Set    `tfsdk:"member"`
	MemberGroup    types.Set    `tfsdk:"member_group"`
}

func (d *orgCollectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_collection"
}

func (d *orgCollectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema_definition.OrgCollectionDataSourceSchema()
}

func (d *orgCollectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *orgCollectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg orgCollectionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(d.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeOrganizationID: cfg.OrganizationID.ValueString(),
		schema_definition.AttributeFilterSearch:   cfg.Search.ValueString(),
	})
	attr.SetId(cfg.ID.ValueString())

	var (
		obj *models.OrgCollection
		err error
	)
	if cfg.ID.ValueString() != "" {
		obj, err = bwClient.GetOrganizationCollection(ctx, transformation.OrganizationCollectionToObject(ctx, attr))
	} else {
		obj, err = bwClient.FindOrganizationCollection(ctx, transformation.ListOptionsFromData(attr)...)
	}
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.OrganizationCollectionObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	vals := attr.Values()
	cfg.ID = types.StringValue(attr.Id())
	cfg.Name = mapStr(vals[schema_definition.AttributeName])
	cfg.OrganizationID = mapStr(vals[schema_definition.AttributeOrganizationID])
	cfg.Member = membershipDataToSet(vals[schema_definition.AttributeMember])
	cfg.MemberGroup = membershipDataToSet(vals[schema_definition.AttributeMemberGroup])
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
