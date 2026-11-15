package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

var (
	_ ephemeral.EphemeralResource              = &itemEphemeralResource[itemSecureNoteDataSourceModel]{}
	_ ephemeral.EphemeralResourceWithConfigure = &itemEphemeralResource[itemSecureNoteDataSourceModel]{}
)

func NewItemSecureNoteEphemeralResource() ephemeral.EphemeralResource {
	return &itemEphemeralResource[itemSecureNoteDataSourceModel]{
		typeNameSuffix: "_item_secure_note",
		itemType:       models.ItemTypeSecureNote,
		schema:         schema_definition.SecureNoteEphemeralResourceSchema,
		prepareAttr:    prepareItemSecureNoteAttr,
		resultToState:  itemSecureNoteResultToState,
	}
}
