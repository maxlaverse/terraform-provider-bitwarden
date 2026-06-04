package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opOrganizationGroupCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwClient.CreateOrganizationGroup, transformation.OrganizationGroupToObject, transformation.OrganizationGroupObjectToSchema))
}

func opOrganizationGroupDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwClient.DeleteOrganizationGroup), transformation.OrganizationGroupToObject, transformation.OrganizationGroupObjectToSchema))
}

func opOrganizationGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid ID specified, should be in the format <organization_id>/<group_id>: '%s'", d.Id())
	}
	d.SetId(split[1])

	err := d.Set(schema_definition.AttributeOrganizationID, split[0])
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func opOrganizationGroupRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		orgId := d.Get(schema_definition.AttributeOrganizationID).(string)

		// Per schema, if the ID is not provided then the name has.
		nameFilter := d.Get(schema_definition.AttributeFilterName).(string)

		obj, err := bwClient.FindOrganizationGroup(ctx, bitwarden.WithOrganizationID(orgId), bitwarden.WithSearch(nameFilter))
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.FromErr(transformation.OrganizationGroupObjectToSchema(ctx, obj, d))
	}

	return diag.FromErr(applyOperation(ctx, d, bwClient.GetOrganizationGroup, transformation.OrganizationGroupToObject, transformation.OrganizationGroupObjectToSchema))
}

func opOrganizationGroupReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwClient.GetOrganizationGroup, transformation.OrganizationGroupToObject, transformation.OrganizationGroupObjectToSchema))
}

func opOrganizationGroupUpdate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwClient.EditOrganizationGroup, transformation.OrganizationGroupToObject, transformation.OrganizationGroupObjectToSchema))
}
