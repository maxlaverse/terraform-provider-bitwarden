package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

var (
	_ ephemeral.EphemeralResource              = &attachmentEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &attachmentEphemeralResource{}
)

type attachmentEphemeralResource struct {
	clients *ProviderClients
}

func NewAttachmentEphemeralResource() ephemeral.EphemeralResource {
	return &attachmentEphemeralResource{}
}

func (r *attachmentEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attachment"
}

func (r *attachmentEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema_definition.AttachmentEphemeralResourceSchema()
}

func (r *attachmentEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	r.clients = clients
}

func (r *attachmentEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var cfg attachmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	content, err := bwClient.GetAttachment(ctx, cfg.ItemID.ValueString(), cfg.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	cfg.Content = types.StringValue(string(content))
	resp.Diagnostics.Append(resp.Result.Set(ctx, cfg)...)
}
