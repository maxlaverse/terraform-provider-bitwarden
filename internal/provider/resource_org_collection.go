package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func resourceOrgCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an organization collection.",

		CreateContext: resourceOrgCollectionCreate,
		ReadContext:   objectReadIgnoreMissing,
		UpdateContext: objectUpdate,
		DeleteContext: objectDelete,
		Importer:      importOrgCollectionResource(),

		Schema: orgCollectionSchema(Resource),
	}
}

func resourceOrgCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := d.Set(attributeObject, bw.ObjectTypeOrgCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	return objectCreate(ctx, d, meta)
}

func importOrgCollectionResource() *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			split := strings.Split(d.Id(), "/")
			if len(split) != 2 {
				return nil, fmt.Errorf("invalid ID specified, should be in the format <organization_id>/<collection_id>: '%s'", d.Id())
			}
			d.SetId(split[1])
			d.Set(attributeOrganizationID, split[0])
			err := d.Set(attributeObject, bw.ObjectTypeOrgCollection)
			if err != nil {
				return nil, err
			}

			return []*schema.ResourceData{d}, nil
		},
	}
}
