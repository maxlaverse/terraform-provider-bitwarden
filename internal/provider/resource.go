package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func createResource(attrObject models.ObjectType, attrType models.ItemType) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		err := d.Set(attributeObject, attrObject)
		if err != nil {
			return diag.FromErr(err)
		}
		err = d.Set(attributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}
		return objectCreate(ctx, d, meta)
	}
}

func importItemResource(attrObject models.ObjectType, attrType models.ItemType) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			d.SetId(d.Id())
			err := d.Set(attributeObject, attrObject)
			if err != nil {
				return nil, err
			}
			err = d.Set(attributeType, attrType)
			if err != nil {
				return nil, err
			}
			return []*schema.ResourceData{d}, nil
		},
	}
}
