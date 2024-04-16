package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func readDataSource(attrObject bw.ObjectType, attrType bw.ItemType) schema.ReadContextFunc {
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

func listOptionsFromData(d *schema.ResourceData) []bw.ListObjectsOption {
	filters := []bw.ListObjectsOption{}

	filterMap := map[string]bw.ListObjectsOptionGenerator{
		attributeFilterSearch:         bw.WithSearch,
		attributeFilterCollectionId:   bw.WithCollectionID,
		attributeFilterFolderID:       bw.WithFolderID,
		attributeFilterOrganizationID: bw.WithOrganizationID,
		attributeFilterURL:            bw.WithUrl,
	}

	for attribute, optionFunc := range filterMap {
		v, ok := d.GetOk(attribute)
		if !ok {
			continue
		}

		if v, ok := v.(string); ok && len(v) > 0 {
			filters = append(filters, optionFunc(v))
		}
	}
	return filters
}
