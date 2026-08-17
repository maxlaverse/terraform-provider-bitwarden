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
	_ datasource.DataSource              = &secretDataSource{}
	_ datasource.DataSourceWithConfigure = &secretDataSource{}
)

type secretDataSource struct {
	clients *ProviderClients
}

func NewSecretDataSource() datasource.DataSource {
	return &secretDataSource{}
}

type secretDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	Note           types.String `tfsdk:"note"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectID      types.String `tfsdk:"project_id"`
}

func (d *secretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (d *secretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema_definition.SecretDataSourceSchema()
}

func (d *secretDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *secretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg secretDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwsClient, ok := requireSecretsManager(d.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := transformation.NewMapData(map[string]interface{}{
		schema_definition.AttributeKey:            cfg.Key.ValueString(),
		schema_definition.AttributeOrganizationID: cfg.OrganizationID.ValueString(),
	})
	attr.SetId(cfg.ID.ValueString())

	var (
		obj *models.Secret
		err error
	)
	if cfg.ID.ValueString() != "" {
		obj, err = bwsClient.GetSecret(ctx, transformation.SecretSchemaToObject(ctx, attr))
	} else {
		obj, err = bwsClient.GetSecretByKey(ctx, cfg.Key.ValueString())
	}
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.SecretObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	vals := attr.Values()
	cfg.ID = types.StringValue(attr.Id())
	cfg.Key = mapStr(vals[schema_definition.AttributeKey])
	cfg.Value = mapStr(vals[schema_definition.AttributeValue])
	cfg.Note = mapStr(vals[schema_definition.AttributeNote])
	cfg.OrganizationID = mapStr(vals[schema_definition.AttributeOrganizationID])
	cfg.ProjectID = mapStr(vals[schema_definition.AttributeProjectID])
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
