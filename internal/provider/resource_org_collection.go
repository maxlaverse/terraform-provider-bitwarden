package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceOrgCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an organization collection.",

		CreateContext: withPasswordManager(resourceCreateOrgCollection),
		ReadContext:   withPasswordManager(resourceReadObjectIgnoreMissing),
		UpdateContext: withPasswordManager(resourceUpdateObject),
		DeleteContext: withPasswordManager(resourceDeleteObject),
		Importer:      resourceImporter(resourceImportOrgCollection),

		Schema: schema_definition.OrgCollectionSchema(schema_definition.Resource),
	}
}

func resourceCreateOrgCollection(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return diag.FromErr(err)
	}

	return objectCreate(ctx, d, bwClient)
}

func resourceImportOrgCollection(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid ID specified, should be in the format <organization_id>/<collection_id>: '%s'", d.Id())
	}
	d.SetId(split[1])
	d.Set(schema_definition.AttributeOrganizationID, split[0])
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
