package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func NewItemLoginResource() resource.Resource {
	return &itemResource[itemLoginModel]{
		typeNameSuffix: "_item_login",
		itemType:       models.ItemTypeLogin,
		schema:         schema_definition.LoginResourceSchema,
		toData:         loginToData,
		fromData:       loginFromData,
	}
}
