package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

var (
	_ ephemeral.EphemeralResource              = &itemEphemeralResource[itemSSHKeyDataSourceModel]{}
	_ ephemeral.EphemeralResourceWithConfigure = &itemEphemeralResource[itemSSHKeyDataSourceModel]{}
)

func NewItemSSHKeyEphemeralResource() ephemeral.EphemeralResource {
	return &itemEphemeralResource[itemSSHKeyDataSourceModel]{
		typeNameSuffix: "_item_ssh_key",
		itemType:       models.ItemTypeSSHKey,
		schema:         schema_definition.SSHKeyEphemeralResourceSchema,
		prepareAttr:    prepareItemSSHKeyAttr,
		resultToState:  itemSSHKeyResultToState,
	}
}
