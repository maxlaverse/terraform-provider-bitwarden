package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func resourceReadDataSourceItem(attrObject models.ObjectType, attrType models.ItemType) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		err := d.Set(attributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}
		return resourceReadDataSourceObject(attrObject)(ctx, d, meta)
	}
}

func resourceReadDataSourceObject(objType models.ObjectType) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		d.SetId(d.Get(attributeID).(string))
		err := d.Set(attributeObject, objType)
		if err != nil {
			return diag.FromErr(err)
		}
		bwClient, err := getPasswordManager(meta)
		if err != nil {
			return diag.FromErr(err)
		}
		return objectRead(ctx, d, bwClient)
	}
}
