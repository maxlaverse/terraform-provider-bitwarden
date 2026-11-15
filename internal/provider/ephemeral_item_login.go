package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

var (
	_ ephemeral.EphemeralResource              = &itemEphemeralResource[itemLoginDataSourceModel]{}
	_ ephemeral.EphemeralResourceWithConfigure = &itemEphemeralResource[itemLoginDataSourceModel]{}
)

func NewItemLoginEphemeralResource() ephemeral.EphemeralResource {
	return &itemEphemeralResource[itemLoginDataSourceModel]{
		typeNameSuffix: "_item_login",
		itemType:       models.ItemTypeLogin,
		schema:         schema_definition.LoginEphemeralResourceSchema,
		prepareAttr:    prepareItemLoginAttr,
		resultToState:  itemLoginResultToState,
	}
}
