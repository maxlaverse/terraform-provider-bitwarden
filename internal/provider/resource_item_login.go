package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type itemLoginModel struct {
	itemVaultModel
	Password types.String `tfsdk:"password"`
	Username types.String `tfsdk:"username"`
	Totp     types.String `tfsdk:"totp"`
	URI      types.List   `tfsdk:"uri"`
}

// itemLoginResourceModel adds the resource-only write-only attributes to the shared login model.
type itemLoginResourceModel struct {
	itemLoginModel
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64  `tfsdk:"password_wo_version"`
}

func loginToData(ctx context.Context, model itemLoginModel) *transformation.MapData {
	values := model.itemVaultModel.toDataMap(ctx)
	values[schema_definition.AttributeLoginPassword] = model.Password.ValueString()
	values[schema_definition.AttributeLoginUsername] = model.Username.ValueString()
	values[schema_definition.AttributeLoginTotp] = model.Totp.ValueString()
	values[schema_definition.AttributeLoginURIs] = uriListToData(model.URI)
	return newItemMapData(model.ID.ValueString(), values)
}

func loginFromData(attr *transformation.MapData) itemLoginModel {
	values := attr.Values()
	return itemLoginModel{
		itemVaultModel: itemVaultFromValues(attr.Id(), values),
		Password:       mapStr(values[schema_definition.AttributeLoginPassword]),
		Username:       mapStr(values[schema_definition.AttributeLoginUsername]),
		Totp:           mapStr(values[schema_definition.AttributeLoginTotp]),
		URI:            uriDataToList(values[schema_definition.AttributeLoginURIs]),
	}
}

// loginOverlayConfig injects the write-only password, readable only from config, into the API payload.
func loginResourceToData(ctx context.Context, m itemLoginResourceModel) *transformation.MapData {
	return loginToData(ctx, m.itemLoginModel)
}

func loginResourceFromData(attr *transformation.MapData) itemLoginResourceModel {
	return itemLoginResourceModel{itemLoginModel: loginFromData(attr)}
}

func loginOverlayConfig(ctx context.Context, cfg tfsdk.Config, attr *transformation.MapData) diag.Diagnostics {
	var pw types.String
	diags := cfg.GetAttribute(ctx, path.Root(schema_definition.AttributeLoginPasswordWO), &pw)
	if diags.HasError() || pw.IsNull() || pw.IsUnknown() {
		return diags
	}
	attr.Values()[schema_definition.AttributeLoginPassword] = pw.ValueString()
	return diags
}

// loginFinalize keeps the password out of state when write-only mode is on and carries the version over.
func loginFinalize(out *itemLoginResourceModel, prior itemLoginResourceModel) {
	out.PasswordWOVersion = prior.PasswordWOVersion
	if !prior.PasswordWOVersion.IsNull() {
		out.Password = types.StringNull()
	}
}

func NewItemLoginResource() resource.Resource {
	return &itemResource[itemLoginResourceModel]{
		typeNameSuffix: "_item_login",
		itemType:       models.ItemTypeLogin,
		schema:         schema_definition.LoginResourceSchema,
		toData:         loginResourceToData,
		fromData:       loginResourceFromData,
		overlayConfig:  loginOverlayConfig,
		finalize:       loginFinalize,
	}
}
