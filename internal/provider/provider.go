package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	provschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

// Ensure the provider satisfies the framework interface.
var _ provider.Provider = &bitwardenProvider{}

type bitwardenProvider struct {
	version string
}

// New returns a Plugin Framework provider factory.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &bitwardenProvider{version: version}
	}
}

type experimentalModel struct {
	EmbeddedClient                    types.Bool `tfsdk:"embedded_client"`
	DisableSyncAfterWriteVerification types.Bool `tfsdk:"disable_sync_after_write_verification"`
}

type bitwardenProviderModel struct {
	MasterPassword       types.String `tfsdk:"master_password"`
	SessionKey           types.String `tfsdk:"session_key"`
	ClientID             types.String `tfsdk:"client_id"`
	ClientSecret         types.String `tfsdk:"client_secret"`
	AccessToken          types.String `tfsdk:"access_token"`
	Server               types.String `tfsdk:"server"`
	Email                types.String `tfsdk:"email"`
	VaultPath            types.String `tfsdk:"vault_path"`
	ExtraCACerts         types.String `tfsdk:"extra_ca_certs"`
	ClientImplementation types.String `tfsdk:"client_implementation"`
	Experimental         types.Set    `tfsdk:"experimental"`
}

func (p *bitwardenProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bitwarden"
	resp.Version = p.version
}

func (p *bitwardenProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	// Provider schema must match NewSDK for terraform-plugin-mux. Keep flags
	// (sensitive, markdown descriptions) aligned with the SDKv2 provider schema.
	resp.Schema = provschema.Schema{
		Attributes: map[string]provschema.Attribute{
			// Credential attributes (cross-field rules enforced in validateProviderConfig)
			schema_definition.AttributeMasterPassword: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionMasterPassword,
				Optional:            true,
				Sensitive:           true,
			},
			schema_definition.AttributeSessionKey: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionSessionKey,
				Optional:            true,
				Sensitive:           true,
			},
			schema_definition.AttributeClientID: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionClientID,
				Optional:            true,
			},
			schema_definition.AttributeClientSecret: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionClientSecret,
				Optional:            true,
				Sensitive:           true,
			},
			schema_definition.AttributeBwsAccessToken: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionBwsAccessToken,
				Optional:            true,
				Sensitive:           true,
			},

			// Standalone attributes
			schema_definition.AttributeServer: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionServer,
				Optional:            true,
			},
			schema_definition.AttributeProviderEmail: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionProviderEmail,
				Optional:            true,
			},
			schema_definition.AttributeVaultPath: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionVaultPath,
				Optional:            true,
			},
			schema_definition.AttributeExtraCACertsPath: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionExtraCACertsPath,
				Optional:            true,
			},
			schema_definition.AttributeClientImplementation: provschema.StringAttribute{
				MarkdownDescription: schema_definition.DescriptionClientImplementation,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(schema_definition.ClientImplementationCLI, schema_definition.ClientImplementationEmbedded),
				},
			},
		},
		Blocks: map[string]provschema.Block{
			// Experimental
			schema_definition.AttributeExperimental: provschema.SetNestedBlock{
				MarkdownDescription: schema_definition.DescriptionExperimental,
				NestedObject: provschema.NestedBlockObject{
					Attributes: map[string]provschema.Attribute{
						schema_definition.AttributeExperimentalEmbeddedClient: provschema.BoolAttribute{
							MarkdownDescription: schema_definition.DescriptionExperimentalEmbeddedClient,
							Optional:            true,
							DeprecationMessage:  "Use client_implementation = \"embedded\" instead.",
							// Both may be set; experimental.embedded_client wins in getClientImplementation().
							// client_implementation is defaulted there when unset, so a schema-level conflict
							// cannot distinguish an explicit "cli" from the default.
						},
						schema_definition.AttributeExperimentalDisableSyncAfterWriteVerification: provschema.BoolAttribute{
							MarkdownDescription: schema_definition.DescriptionExperimentalDisableSyncAfterWriteVerification,
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (p *bitwardenProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// When Framework still registers nothing, SDKv2 alone owns login — skip here
	// to avoid parking unused clients. Once any Framework type is registered,
	// Configure offers clients for the SDKv2 mux half to take.
	if !p.ownsManagedResources(ctx) {
		return
	}

	var model bitwardenProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := applyProviderConfigEnvDefaults(providerConfig{
		Server:               model.Server.ValueString(),
		Email:                model.Email.ValueString(),
		MasterPassword:       model.MasterPassword.ValueString(),
		SessionKey:           model.SessionKey.ValueString(),
		ClientID:             model.ClientID.ValueString(),
		ClientSecret:         model.ClientSecret.ValueString(),
		AccessToken:          model.AccessToken.ValueString(),
		VaultPath:            model.VaultPath.ValueString(),
		ExtraCACertsPath:     model.ExtraCACerts.ValueString(),
		ClientImplementation: model.ClientImplementation.ValueString(),
	})

	if !model.Experimental.IsNull() && !model.Experimental.IsUnknown() {
		var experimental []experimentalModel
		resp.Diagnostics.Append(model.Experimental.ElementsAs(ctx, &experimental, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(experimental) > 0 {
			cfg.ExperimentalEmbeddedClient = experimental[0].EmbeddedClient.ValueBool()
			cfg.ExperimentalDisableSyncAfterWriteVerification = experimental[0].DisableSyncAfterWriteVerification.ValueBool()
		}
	}

	if err := validateProviderConfig(cfg); err != nil {
		resp.Diagnostics.AddError("Missing required argument", err.Error())
		return
	}

	// Mux configures NewSDK next; offer clients so SDKv2 can reuse this login
	// for the same ConfigureProvider RPC instead of authenticating twice.
	clients, err := configureClientsOffer(ctx, p.version, cfg)
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.ResourceData = clients
	resp.DataSourceData = clients
}

func (p *bitwardenProvider) ownsManagedResources(ctx context.Context) bool {
	// True when Framework registers at least one resource or data source and
	// must supply ProviderData via Configure. While false, NewSDK owns login.
	return len(p.Resources(ctx)) > 0 || len(p.DataSources(ctx)) > 0
}

func (p *bitwardenProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFolderResource,
		NewProjectResource,
	}
}

func (p *bitwardenProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFolderDataSource,
		NewOrganizationDataSource,
		NewOrgGroupDataSource,
		NewOrgMemberDataSource,
		NewProjectDataSource,
	}
}
