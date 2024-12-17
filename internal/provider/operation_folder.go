package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func opFolderCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}

	return objectCreate(ctx, d, bwClient)
}

func opFolderImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeFolder)
	if err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
