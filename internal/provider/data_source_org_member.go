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
	_ datasource.DataSource              = &orgMemberDataSource{}
	_ datasource.DataSourceWithConfigure = &orgMemberDataSource{}
)

type orgMemberDataSource struct {
	clients *ProviderClients
}

func NewOrgMemberDataSource() datasource.DataSource {
	return &orgMemberDataSource{}
}

type orgMemberDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Email          types.String `tfsdk:"email"`
}

func (d *orgMemberDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_member"
}

func (d *orgMemberDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema_definition.OrgMemberDataSourceSchema()
}

func (d *orgMemberDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *orgMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg orgMemberDataSourceModel
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
		schema_definition.AttributeEmail:          cfg.Email.ValueString(),
	})
	attr.SetId(cfg.ID.ValueString())

	var (
		obj *models.OrgMember
		err error
	)
	if cfg.ID.ValueString() != "" {
		obj, err = bwClient.GetOrganizationMember(ctx, transformation.OrganizationMemberToObject(ctx, attr))
	} else {
		obj, err = bwClient.FindOrganizationMember(ctx,
			bitwarden.WithOrganizationID(cfg.OrganizationID.ValueString()),
			bitwarden.WithSearch(cfg.Email.ValueString()),
		)
	}
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	if err = transformation.OrganizationMemberObjectToSchema(ctx, obj, attr); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	cfg.ID = types.StringValue(attr.Id())
	cfg.Name = mapStr(attr.Values()[schema_definition.AttributeName])
	cfg.Email = mapStr(attr.Values()[schema_definition.AttributeEmail])
	cfg.OrganizationID = mapStr(attr.Values()[schema_definition.AttributeOrganizationID])
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
