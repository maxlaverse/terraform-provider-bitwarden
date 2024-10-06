package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func resourceFolder() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a folder.",

		CreateContext: withPasswordManager(resourceCreateFolder),
		ReadContext:   withPasswordManager(resourceReadObjectIgnoreMissing),
		UpdateContext: withPasswordManager(resourceUpdateObject),
		DeleteContext: withPasswordManager(resourceDeleteObject),
		Importer:      resourceImporter(resourceImportFolder),

		Schema: folderSchema(Resource),
	}
}

func resourceCreateFolder(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(attributeObject, models.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}

	return objectCreate(ctx, d, bwClient)
}

func resourceImportFolder(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	err := d.Set(attributeObject, models.ObjectTypeFolder)
	if err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
