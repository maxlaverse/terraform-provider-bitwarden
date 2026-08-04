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
	_ datasource.DataSource              = &folderDataSource{}
	_ datasource.DataSourceWithConfigure = &folderDataSource{}
)

type folderDataSource struct {
	clients *ProviderClients
}

func NewFolderDataSource() datasource.DataSource {
	return &folderDataSource{}
}

type folderDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	FilterCollectionID   types.String `tfsdk:"filter_collection_id"`
	FilterOrganizationID types.String `tfsdk:"filter_organization_id"`
	Search               types.String `tfsdk:"search"`
}

func (d *folderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

func (d *folderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema_definition.FolderDataSourceSchema()
}

func (d *folderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *folderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg folderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(d.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeFilterCollectionId:   cfg.FilterCollectionID.ValueString(),
		schema_definition.AttributeFilterOrganizationID: cfg.FilterOrganizationID.ValueString(),
		schema_definition.AttributeFilterSearch:         cfg.Search.ValueString(),
	})
	attr.SetId(cfg.ID.ValueString())

	var (
		obj *models.Folder
		err error
	)
	if cfg.ID.ValueString() != "" {
		obj, err = bwClient.GetFolder(ctx, transformation.SchemaToFolderObject(ctx, attr))
	} else {
		obj, err = bwClient.FindFolder(ctx, transformation.ListOptionsFromData(attr)...)
	}
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.FolderObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	cfg.ID = types.StringValue(attr.Id())
	cfg.Name = mapStr(attr.Values()[schema_definition.AttributeName])
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
