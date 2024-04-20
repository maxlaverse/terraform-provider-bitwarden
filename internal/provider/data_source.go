package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func readDataSourceItem(attrObject bw.ObjectType, attrType bw.ItemType) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		d.SetId(d.Get(attributeID).(string))
		err := d.Set(attributeObject, attrObject)
		if err != nil {
			return diag.FromErr(err)
		}
		err = d.Set(attributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}
		return objectRead(ctx, d, meta)
	}
}

func readDataSourceObject(objType bw.ObjectType) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		d.SetId(d.Get(attributeID).(string))
		err := d.Set(attributeObject, objType)
		if err != nil {
			return diag.FromErr(err)
		}
		return objectRead(ctx, d, meta)
	}
}
