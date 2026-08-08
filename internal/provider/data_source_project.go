package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

type projectDataSource struct {
	clients *ProviderClients
}

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema_definition.ProjectDataSourceSchema()
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(d.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeOrganizationID: cfg.OrganizationID.ValueString(),
	})
	attr.SetId(cfg.ID.ValueString())

	obj, err := bwsClient.GetProject(ctx, transformation.ProjectSchemaToObject(ctx, attr))
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.ProjectObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	cfg.ID = types.StringValue(attr.Id())
	cfg.Name = mapStr(attr.Values()[schema_definition.AttributeName])
	cfg.OrganizationID = mapStr(attr.Values()[schema_definition.AttributeOrganizationID])
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
