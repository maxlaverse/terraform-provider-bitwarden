package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type itemSSHKeyModel struct {
	itemCommonModel
	PrivateKey     types.String `tfsdk:"private_key"`
	PublicKey      types.String `tfsdk:"public_key"`
	KeyFingerprint types.String `tfsdk:"key_fingerprint"`
}

func sshKeyToData(ctx context.Context, model itemSSHKeyModel) *transformation.MapData {
	values := model.itemCommonModel.toDataMap(ctx)
	values[schema_definition.AttributeSSHKeyPrivateKey] = model.PrivateKey.ValueString()
	values[schema_definition.AttributeSSHKeyPublicKey] = model.PublicKey.ValueString()
	values[schema_definition.AttributeSSHKeyKeyFingerprint] = model.KeyFingerprint.ValueString()
	return newItemMapData(model.ID.ValueString(), values)
}

func sshKeyFromData(attr *transformation.MapData) itemSSHKeyModel {
	values := attr.Values()
	return itemSSHKeyModel{
		itemCommonModel: itemCommonFromValues(attr.Id(), values),
		PrivateKey:      mapStr(values[schema_definition.AttributeSSHKeyPrivateKey]),
		PublicKey:       mapStr(values[schema_definition.AttributeSSHKeyPublicKey]),
		KeyFingerprint:  mapStr(values[schema_definition.AttributeSSHKeyKeyFingerprint]),
	}
}

func NewItemSSHKeyResource() resource.Resource {
	return &itemResource[itemSSHKeyModel]{
		typeNameSuffix: "_item_ssh_key",
		itemType:       models.ItemTypeSSHKey,
		schema:         schema_definition.SSHKeyResourceSchema,
		toData:         sshKeyToData,
		fromData:       sshKeyFromData,
	}
}
