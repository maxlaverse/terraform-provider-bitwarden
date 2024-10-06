package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func resourceOrgCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an organization collection.",

		CreateContext: resourceCreateOrgCollection,
		ReadContext:   resourceReadObjectIgnoreMissing,
		UpdateContext: resourceUpdateObject,
		DeleteContext: resourceDeleteObject,
		Importer:      resourceImportOrgCollection(),

		Schema: orgCollectionSchema(Resource),
	}
}

func resourceCreateOrgCollection(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := d.Set(attributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	bwClient, err := getPasswordManager(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return objectCreate(ctx, d, bwClient)
}

func resourceImportOrgCollection() *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			split := strings.Split(d.Id(), "/")
			if len(split) != 2 {
				return nil, fmt.Errorf("invalid ID specified, should be in the format <organization_id>/<collection_id>: '%s'", d.Id())
			}
			d.SetId(split[1])
			d.Set(attributeOrganizationID, split[0])
			err := d.Set(attributeObject, models.ObjectTypeOrgCollection)
			if err != nil {
				return nil, err
			}

			return []*schema.ResourceData{d}, nil
		},
	}
}
