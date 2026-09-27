package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

var (
	_ datasource.DataSource              = &attachmentDataSource{}
	_ datasource.DataSourceWithConfigure = &attachmentDataSource{}
)

type attachmentDataSource struct {
	clients *ProviderClients
}

func NewAttachmentDataSource() datasource.DataSource {
	return &attachmentDataSource{}
}

type attachmentDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	ItemID  types.String `tfsdk:"item_id"`
	Content types.String `tfsdk:"content"`
}

func (d *attachmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attachment"
}

func (d *attachmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema_definition.AttachmentDataSourceSchema()
}

func (d *attachmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *attachmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg attachmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(d.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	content, err := bwClient.GetAttachment(ctx, cfg.ItemID.ValueString(), cfg.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	cfg.Content = types.StringValue(string(content))
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
