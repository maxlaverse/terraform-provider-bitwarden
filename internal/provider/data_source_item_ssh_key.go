package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type itemSSHKeyDataSourceModel struct {
	itemSSHKeyModel
	itemFilterModel
}

func NewItemSSHKeyDataSource() datasource.DataSource {
	return &itemDataSource[itemSSHKeyDataSourceModel]{
		typeNameSuffix: "_item_ssh_key",
		itemType:       models.ItemTypeSSHKey,
		schema:         schema_definition.SSHKeyDataSourceSchema,
		prepareAttr: func(ctx context.Context, cfg itemSSHKeyDataSourceModel) (*transformation.MapData, string) {
			attr := sshKeyToData(ctx, cfg.itemSSHKeyModel)
			cfg.itemFilterModel.applyTo(attr)
			return attr, cfg.ID.ValueString()
		},
		resultToState: func(attr *transformation.MapData, cfg itemSSHKeyDataSourceModel) itemSSHKeyDataSourceModel {
			return itemSSHKeyDataSourceModel{
				itemSSHKeyModel: sshKeyFromData(attr),
				itemFilterModel: cfg.itemFilterModel,
			}
		},
	}
}
