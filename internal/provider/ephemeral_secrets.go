package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	IDs     types.Set `tfsdk:"ids"`
	Secrets types.Map `tfsdk:"secrets"`
}

type batchSecretModel struct {
	Key   string `tfsdk:"key"`
	Value string `tfsdk:"value"`
	Note  string `tfsdk:"note"`
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

	values := make(map[string]batchSecretModel, len(ids))
	if len(ids) > 0 {
		secrets, err := client.GetSecretsByIDs(ctx, ids)
		if err != nil {
			resp.Diagnostics.AddError("Unable to read secrets", ephemeralSecretError(err))
			return
		}
		for _, secret := range secrets {
			values[secret.ID] = batchSecretModel{Key: secret.Key, Value: secret.Value, Note: secret.Note}
		}
	}

	objectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"key": types.StringType, "value": types.StringType, "note": types.StringType,
	}}
	valueMap, diagnostics := types.MapValueFrom(ctx, objectType, values)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Secrets = valueMap
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
