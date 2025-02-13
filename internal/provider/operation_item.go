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
		err := d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}

		return diag.FromErr(applyOperation(ctx, d, bwClient.CreateItem, transformation.ItemSchemaToObject, transformation.ItemObjectToSchema))
	}
}

func opItemDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwClient.DeleteItem), transformation.ItemSchemaToObject, transformation.ItemObjectToSchema))
}

func opItemImport(attrType models.ItemType) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		d.SetId(d.Id())
		err := d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return nil, err
		}
		return []*schema.ResourceData{d}, nil
	}
}

func opItemRead(attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		err := d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(d.Get(schema_definition.AttributeID).(string))
		if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
			return diag.FromErr(searchOperation(ctx, d, bwClient.FindItem, transformation.ItemObjectToSchema))
		}
		return diag.FromErr(applyOperation(ctx, d, bwClient.GetItem, transformation.ItemSchemaToObject, transformation.ItemObjectToSchema))
	}
}

func opItemReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwClient.GetItem, transformation.ItemSchemaToObject, transformation.ItemObjectToSchema))
}

func opItemUpdate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwClient.EditItem, transformation.ItemSchemaToObject, transformation.ItemObjectToSchema))
}
