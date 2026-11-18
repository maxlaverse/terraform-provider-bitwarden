package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
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

// itemLoginResourceModel extends itemLoginModel with the write-only
// attributes, which only the resource (not the data source or ephemeral
// resource) exposes.
type itemLoginResourceModel struct {
	itemLoginModel
	PasswordWO types.String `tfsdk:"password_wo"`
	UsernameWO types.String `tfsdk:"username_wo"`
	WOVersion  types.Int64  `tfsdk:"wo_version"`
}

func loginResourceToData(ctx context.Context, model itemLoginResourceModel) *transformation.MapData {
	attr := loginToData(ctx, model.itemLoginModel)

	// password/username are never known from the typed model alone once a
	// write-only variant is in play (Terraform always nulls password_wo /
	// username_wo out of plan and state). Default to "write-only in use"
	// whenever the plain attribute is currently null; loginPrepareWriteOnly
	// then corrects this with the authoritative answer taken from config on
	// Create/Update. Read has no such hook, but by then a prior Create/Update
	// has already nulled out the plain attribute in state if it's write-only
	// sourced, so the same check is accurate there too -- except right after
	// Import, where every attribute (including the Required Name) is null in
	// the skeleton state handed to the following Read. There's no way to tell
	// whether such an item was write-only sourced, so default to "no" (mirror
	// normally), matching every other write-only Terraform resource.
	isFreshImport := model.Name.IsNull()
	_ = attr.Set(schema_definition.AttributeLoginPasswordWO, !isFreshImport && model.Password.IsNull())
	_ = attr.Set(schema_definition.AttributeLoginUsernameWO, !isFreshImport && model.Username.IsNull())
	if model.WOVersion.IsNull() {
		_ = attr.Set(schema_definition.AttributeLoginWOVersion, nil)
	} else {
		_ = attr.Set(schema_definition.AttributeLoginWOVersion, model.WOVersion.ValueInt64())
	}
	return attr
}

// loginResourceFromData intentionally leaves PasswordWO/UsernameWO unset
// (null): write-only attribute values must never be persisted to state.
func loginResourceFromData(attr *transformation.MapData) itemLoginResourceModel {
	return itemLoginResourceModel{
		itemLoginModel: loginFromData(attr),
		WOVersion:      mapInt64(attr.Values()[schema_definition.AttributeLoginWOVersion]),
	}
}

// loginPrepareWriteOnly resolves password_wo/username_wo into attr right
// before it's sent to the API. Their value is only ever available from the
// raw config (Terraform always nulls them out of plan and state), so it's
// read from cfg rather than the plan model.
//
// On update, the write-only value is only re-applied when wo_version has
// actually changed since the last apply (state is nil on create, where it's
// always applied). Otherwise -- e.g. another attribute changed, or
// password_wo/username_wo are fed by an ephemeral resource that produces a
// fresh value on every apply -- the vault credentials must be left exactly
// as they are; since they're never in state, that requires fetching the
// item's current value from the backend rather than just leaving attr alone.
func loginPrepareWriteOnly(ctx context.Context, bwClient bitwarden.PasswordManager, cfg tfsdk.Config, state *itemLoginResourceModel, plan itemLoginResourceModel, attr *transformation.MapData, diags *diag.Diagnostics) {
	var passwordWO, usernameWO types.String
	diags.Append(cfg.GetAttribute(ctx, path.Root(schema_definition.AttributeLoginPasswordWO), &passwordWO)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root(schema_definition.AttributeLoginUsernameWO), &usernameWO)...)
	if diags.HasError() {
		return
	}

	usingPasswordWO := !passwordWO.IsNull()
	usingUsernameWO := !usernameWO.IsNull()
	_ = attr.Set(schema_definition.AttributeLoginPasswordWO, usingPasswordWO)
	_ = attr.Set(schema_definition.AttributeLoginUsernameWO, usingUsernameWO)

	if !usingPasswordWO && !usingUsernameWO {
		return
	}

	rotate := state == nil || !state.WOVersion.Equal(plan.WOVersion)
	if rotate {
		if usingPasswordWO {
			_ = attr.Set(schema_definition.AttributeLoginPassword, passwordWO.ValueString())
		}
		if usingUsernameWO {
			_ = attr.Set(schema_definition.AttributeLoginUsername, usernameWO.ValueString())
		}
		return
	}

	obj, err := bwClient.GetItem(ctx, models.Item{ID: plan.ID.ValueString(), Object: models.ObjectTypeItem})
	if err != nil {
		diags.AddError("Reading current login before update", err.Error())
		return
	}
	if usingPasswordWO {
		_ = attr.Set(schema_definition.AttributeLoginPassword, obj.Login.Password)
	}
	if usingUsernameWO {
		_ = attr.Set(schema_definition.AttributeLoginUsername, obj.Login.Username)
	}
}

func NewItemLoginResource() resource.Resource {
	return &itemResource[itemLoginResourceModel]{
		typeNameSuffix:   "_item_login",
		itemType:         models.ItemTypeLogin,
		schema:           schema_definition.LoginResourceSchema,
		toData:           loginResourceToData,
		fromData:         loginResourceFromData,
		prepareWriteOnly: loginPrepareWriteOnly,
	}
}
