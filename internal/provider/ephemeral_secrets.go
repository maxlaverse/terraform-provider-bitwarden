package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

var (
	_ ephemeral.EphemeralResource              = &secretsEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &secretsEphemeralResource{}
)

type secretsEphemeralResource struct {
	clients *ProviderClients
}

type secretsEphemeralResourceModel struct {
	IDs    types.Set `tfsdk:"ids"`
	Values types.Map `tfsdk:"values"`
}

func NewSecretsEphemeralResource() ephemeral.EphemeralResource {
	return &secretsEphemeralResource{}
}

func (r *secretsEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secrets"
}

func (r *secretsEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema_definition.SecretsEphemeralResourceSchema()
}

func (r *secretsEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if ok {
		r.clients = clients
	}
}

func (r *secretsEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data secretsEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var ids []string
	resp.Diagnostics.Append(data.IDs.ElementsAs(ctx, &ids, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	values := make(map[string]string, len(ids))
	defer clear(values)
	if len(ids) > 0 {
		secrets, err := client.GetSecretsByIDs(ctx, ids)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read secrets", ephemeralSecretError(err))
			return
		}
		defer clear(secrets)
		for _, secret := range secrets {
			values[secret.ID] = secret.Value
		}
	}

	valueMap, diagnostics := types.MapValueFrom(ctx, types.StringType, values)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Values = valueMap
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
