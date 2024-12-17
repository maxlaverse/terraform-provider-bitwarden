package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceCreateObject(attrObject models.ObjectType, attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		err := d.Set(schema_definition.AttributeObject, attrObject)
		if err != nil {
			return diag.FromErr(err)
		}
		err = d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}

		return objectCreate(ctx, d, bwClient)
	}
}

func resourceReadObjectIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := objectOperation(ctx, d, func(ctx context.Context, secret models.Object) (*models.Object, error) {
		return bwClient.GetObject(ctx, secret)
	})

	if errors.Is(err, models.ErrObjectNotFound) {
		d.SetId("")
		tflog.Warn(ctx, "Object not found, removing from state")
		return diag.Diagnostics{}
	}

	if _, exists := d.GetOk(schema_definition.AttributeDeletedDate); exists {
		d.SetId("")
		tflog.Warn(ctx, "Object was soft deleted, removing from state")
		return diag.Diagnostics{}
	}

	return diag.FromErr(err)
}

func resourceUpdateObject(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(objectOperation(ctx, d, bwClient.EditObject))
}

func resourceDeleteObject(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(objectOperation(ctx, d, func(ctx context.Context, secret models.Object) (*models.Object, error) {
		return nil, bwClient.DeleteObject(ctx, secret)
	}))
}

func resourceImportObject(attrObject models.ObjectType, attrType models.ItemType) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		d.SetId(d.Id())
		err := d.Set(schema_definition.AttributeObject, attrObject)
		if err != nil {
			return nil, err
		}
		err = d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return nil, err
		}
		return []*schema.ResourceData{d}, nil
	}
}

func resourceImporter(stateContext schema.StateContextFunc) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: stateContext,
	}
}
