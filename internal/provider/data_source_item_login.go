package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type itemLoginDataSourceModel struct {
	itemLoginModel
	itemFilterModel
	FilterURL types.String `tfsdk:"filter_url"`
}

func NewItemLoginDataSource() datasource.DataSource {
	return &itemDataSource[itemLoginDataSourceModel]{
		typeNameSuffix: "_item_login",
		itemType:       models.ItemTypeLogin,
		schema:         schema_definition.LoginDataSourceSchema,
		prepareAttr: func(ctx context.Context, cfg itemLoginDataSourceModel) (*transformation.MapData, string) {
			attr := loginToData(ctx, cfg.itemLoginModel)
			cfg.itemFilterModel.applyTo(attr)
			_ = attr.Set(schema_definition.AttributeFilterURL, cfg.FilterURL.ValueString())
			return attr, cfg.ID.ValueString()
		},
		resultToState: func(attr *transformation.MapData, cfg itemLoginDataSourceModel) itemLoginDataSourceModel {
			return itemLoginDataSourceModel{
				itemLoginModel:  loginFromData(attr),
				itemFilterModel: cfg.itemFilterModel,
				FilterURL:       cfg.FilterURL,
			}
		},
	}
}
