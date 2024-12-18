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

func opFolderCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(applyOperation(ctx, d, bwClient.CreateObject, transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}

func opFolderDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwClient.DeleteObject), transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}

func opFolderImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeFolder)
	if err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func opFolderRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		return diag.FromErr(searchOperation(ctx, d, bwClient.ListObjects, transformation.ObjectDataFromStruct))
	}

	return diag.FromErr(applyOperation(ctx, d, bwClient.GetObject, transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}

func opFolderReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwClient.GetObject, transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}

func opFolderUpdate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(applyOperation(ctx, d, bwClient.EditObject, transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}
