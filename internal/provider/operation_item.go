package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opItemCreate(attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		return diag.FromErr(applyOperation(ctx, d, bwClient.CreateItem, transformation.ItemSchemaToObject(attrType), transformation.ItemObjectToSchema))
	}
}

func opItemDelete(attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwClient.DeleteItem), transformation.ItemSchemaToObject(attrType), transformation.ItemObjectToSchema))
	}
}

func opItemImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	return []*schema.ResourceData{d}, nil
}

func opItemRead(attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		d.SetId(d.Get(schema_definition.AttributeID).(string))
		if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
			return diag.FromErr(searchItemOperation(ctx, d, bwClient.FindItem, transformation.ItemObjectToSchema, attrType))
		}
		return diag.FromErr(applyOperation(ctx, d, bwClient.GetItem, transformation.ItemSchemaToObject(attrType), transformation.ItemObjectToSchema))
	}
}

func opItemReadIgnoreMissing(attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		return ignoreMissing(ctx, d, applyOperation(ctx, d, bwClient.GetItem, transformation.ItemSchemaToObject(attrType), transformation.ItemObjectToSchema))
	}
}

func opItemUpdate(attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		return diag.FromErr(applyOperation(ctx, d, bwClient.EditItem, transformation.ItemSchemaToObject(attrType), transformation.ItemObjectToSchema))
	}
}
