package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type itemSecureNoteModel struct {
	itemVaultModel
}

func secureNoteToData(ctx context.Context, model itemSecureNoteModel) *transformation.MapData {
	return newItemMapData(model.ID.ValueString(), model.itemVaultModel.toDataMap(ctx))
}

func secureNoteFromData(attr *transformation.MapData) itemSecureNoteModel {
	return itemSecureNoteModel{
		itemVaultModel: itemVaultFromValues(attr.Id(), attr.Values()),
	}
}

func NewItemSecureNoteResource() resource.Resource {
	return &itemResource[itemSecureNoteModel]{
		typeNameSuffix: "_item_secure_note",
		itemType:       models.ItemTypeSecureNote,
		schema:         schema_definition.SecureNoteResourceSchema,
		toData:         secureNoteToData,
		fromData:       secureNoteFromData,
	}
}
