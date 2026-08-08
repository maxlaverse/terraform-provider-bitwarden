package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type itemSecureNoteDataSourceModel struct {
	itemSecureNoteModel
	itemFilterModel
}

func NewItemSecureNoteDataSource() datasource.DataSource {
	return &itemDataSource[itemSecureNoteDataSourceModel]{
		typeNameSuffix: "_item_secure_note",
		itemType:       models.ItemTypeSecureNote,
		schema:         schema_definition.SecureNoteDataSourceSchema,
		prepareAttr: func(ctx context.Context, cfg itemSecureNoteDataSourceModel) (*transformation.MapData, string) {
			attr := secureNoteToData(ctx, cfg.itemSecureNoteModel)
			cfg.itemFilterModel.applyTo(attr)
			return attr, cfg.ID.ValueString()
		},
		resultToState: func(attr *transformation.MapData, cfg itemSecureNoteDataSourceModel) itemSecureNoteDataSourceModel {
			return itemSecureNoteDataSourceModel{
				itemSecureNoteModel: secureNoteFromData(attr),
				itemFilterModel:     cfg.itemFilterModel,
			}
		},
	}
}
