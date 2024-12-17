package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func opItemRead(attrObject models.ObjectType, attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		err := d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}
		return opObjectRead(attrObject)(ctx, d, bwClient)
	}
}
