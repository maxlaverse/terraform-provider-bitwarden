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
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}
)

type organizationDataSource struct {
	clients *ProviderClients
}

func NewOrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

type organizationDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Search types.String `tfsdk:"search"`
}

func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *organizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema_definition.OrganizationDataSourceSchema()
}

func (d *organizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg organizationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(d.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeFilterSearch: cfg.Search.ValueString(),
	})
	attr.SetId(cfg.ID.ValueString())

	var (
		obj *models.Organization
		err error
	)
	if cfg.ID.ValueString() != "" {
		obj, err = bwClient.GetOrganization(ctx, transformation.OrganizationSchemaToObject(ctx, attr))
	} else {
		obj, err = bwClient.FindOrganization(ctx, transformation.ListOptionsFromData(attr)...)
	}
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.OrganizationObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	cfg.ID = types.StringValue(attr.Id())
	cfg.Name = mapStr(attr.Values()[schema_definition.AttributeName])
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
