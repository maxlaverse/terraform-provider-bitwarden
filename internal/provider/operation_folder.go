package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opFolderCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwClient.CreateFolder, transformation.SchemaToFolderObject, transformation.FolderObjectToSchema))
}

func opFolderDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwClient.DeleteFolder), transformation.SchemaToFolderObject, transformation.FolderObjectToSchema))
}

func opFolderImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	return []*schema.ResourceData{d}, nil
}

func opFolderRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		return diag.FromErr(searchOperation(ctx, d, bwClient.FindFolder, transformation.FolderObjectToSchema))
	}

	return diag.FromErr(applyOperation(ctx, d, bwClient.GetFolder, transformation.SchemaToFolderObject, transformation.FolderObjectToSchema))
}

func opFolderReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwClient.GetFolder, transformation.SchemaToFolderObject, transformation.FolderObjectToSchema))
}

func opFolderUpdate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwClient.EditFolder, transformation.SchemaToFolderObject, transformation.FolderObjectToSchema))
}
