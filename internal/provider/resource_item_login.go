package provider

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	FieldsWO          types.Map    `tfsdk:"fields_wo"`
	FieldsWOVersion   types.Int64  `tfsdk:"fields_wo_version"`
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
	return itemLoginResourceModel{itemLoginModel: loginFromData(attr), FieldsWO: types.MapNull(types.StringType)}
}

func loginOverlayConfig(ctx context.Context, cfg tfsdk.Config, data *transformation.MapData) diag.Diagnostics {
	var pw types.String
	diags := cfg.GetAttribute(ctx, path.Root(schema_definition.AttributeLoginPasswordWO), &pw)
	if diags.HasError() {
		return diags
	}
	if !pw.IsNull() && !pw.IsUnknown() {
		data.Values()[schema_definition.AttributeLoginPassword] = pw.ValueString()
	}

	var fw types.Map
	diags.Append(cfg.GetAttribute(ctx, path.Root(schema_definition.AttributeLoginFieldsWO), &fw)...)
	if diags.HasError() || fw.IsNull() || fw.IsUnknown() {
		return diags
	}
	fields, _ := data.Values()[schema_definition.AttributeField].([]interface{})
	declared := map[string]bool{}
	for _, f := range fields {
		if m, ok := f.(map[string]interface{}); ok {
			if n, ok := m[schema_definition.AttributeFieldName].(string); ok {
				declared[n] = true
			}
		}
	}
	names := make([]string, 0, len(fw.Elements()))
	for n := range fw.Elements() {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		v, ok := fw.Elements()[n].(types.String)
		if !ok || v.IsNull() || v.IsUnknown() {
			continue
		}
		if declared[n] {
			diags.AddError("Duplicate custom field", "Field "+n+" is set in both a `field` block and `fields_wo`.")
			return diags
		}
		fields = append(fields, map[string]interface{}{
			schema_definition.AttributeFieldName:   n,
			schema_definition.AttributeFieldHidden: v.ValueString(),
		})
	}
	data.Values()[schema_definition.AttributeField] = fields
	return diags
}

// loginFinalize keeps the password out of state when write-only mode is on and carries the version over.
func loginFinalize(out *itemLoginResourceModel, prior itemLoginResourceModel) {
	out.PasswordWOVersion = prior.PasswordWOVersion
	if !prior.PasswordWOVersion.IsNull() {
		out.Password = types.StringNull()
	}
	out.FieldsWOVersion = prior.FieldsWOVersion
	if !prior.FieldsWOVersion.IsNull() {
		out.Field = keepDeclaredFields(out.Field, prior.Field)
	}
}

// keepDeclaredFields drops fields the user did not declare in `field` blocks (the write-only ones) so their values stay out of state.
func keepDeclaredFields(got, declared types.List) types.List {
	names := map[string]bool{}
	for _, e := range declared.Elements() {
		if o, ok := e.(types.Object); ok {
			if n, ok := o.Attributes()[schema_definition.AttributeFieldName].(types.String); ok {
				names[n.ValueString()] = true
			}
		}
	}
	kept := []attr.Value{}
	for _, e := range got.Elements() {
		if o, ok := e.(types.Object); ok {
			if n, ok := o.Attributes()[schema_definition.AttributeFieldName].(types.String); ok && names[n.ValueString()] {
				kept = append(kept, e)
			}
		}
	}
	return types.ListValueMust(types.ObjectType{AttrTypes: fieldAttrTypes}, kept)
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
