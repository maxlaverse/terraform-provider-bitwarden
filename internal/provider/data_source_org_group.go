package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

var (
	_ datasource.DataSource              = &orgGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &orgGroupDataSource{}
)

type orgGroupDataSource struct {
	clients *ProviderClients
}

func NewOrgGroupDataSource() datasource.DataSource {
	return &orgGroupDataSource{}
}

type orgGroupDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	FilterName     types.String `tfsdk:"filter_name"`
}

func (d *orgGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_group"
}

func (d *orgGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema_definition.OrgGroupDataSourceSchema()
}

func (d *orgGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *orgGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg orgGroupDataSourceModel
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
		schema_definition.AttributeFilterName:     cfg.FilterName.ValueString(),
	})
	attr.SetId(cfg.ID.ValueString())

	var (
		obj *models.OrgGroup
		err error
	)
	if cfg.ID.ValueString() != "" {
		obj, err = bwClient.GetOrganizationGroup(ctx, transformation.OrganizationGroupToObject(ctx, attr))
	} else {
		obj, err = bwClient.FindOrganizationGroup(ctx,
			bitwarden.WithOrganizationID(cfg.OrganizationID.ValueString()),
			bitwarden.WithSearch(cfg.FilterName.ValueString()),
		)
	}
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.OrganizationGroupObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	cfg.ID = types.StringValue(attr.Id())
	cfg.Name = mapStr(attr.Values()[schema_definition.AttributeName])
	cfg.OrganizationID = mapStr(attr.Values()[schema_definition.AttributeOrganizationID])
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
