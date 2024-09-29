package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func resourceFolder() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a folder.",

		CreateContext: resourceFolderCreate,
		ReadContext:   objectReadIgnoreMissing,
		UpdateContext: objectUpdate,
		DeleteContext: objectDelete,
		Importer:      importFolderResource(),

		Schema: folderSchema(Resource),
	}
}

func resourceFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := d.Set(attributeObject, models.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}
	return objectCreate(ctx, d, meta)
}

func importFolderResource() *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			d.SetId(d.Id())
			err := d.Set(attributeObject, models.ObjectTypeFolder)
			if err != nil {
				return nil, err
			}
			return []*schema.ResourceData{d}, nil
		},
	}
}
