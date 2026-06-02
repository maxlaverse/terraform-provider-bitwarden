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

func opOrganizationMemberCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwClient.CreateOrganizationMember, transformation.OrganizationMemberToObject, transformation.OrganizationMemberObjectToSchema))
}

func opOrganizationMemberRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		orgId := d.Get(schema_definition.AttributeOrganizationID).(string)

		// Per schema, if the ID is not provided then the email has.
		userEmail := d.Get(schema_definition.AttributeEmail).(string)

		obj, err := bwClient.FindOrganizationMember(ctx, bitwarden.WithOrganizationID(orgId), bitwarden.WithSearch(userEmail))
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.FromErr(transformation.OrganizationMemberObjectToSchema(ctx, obj, d))
	}

	return diag.FromErr(applyOperation(ctx, d, bwClient.GetOrganizationMember, transformation.OrganizationMemberToObject, transformation.OrganizationMemberObjectToSchema))
}

func opOrganizationMemberReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwClient.GetOrganizationMember, transformation.OrganizationMemberToObject, transformation.OrganizationMemberObjectToSchema))
}

func opOrganizationMemberDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwClient.DeleteOrganizationMember), transformation.OrganizationMemberToObject, transformation.OrganizationMemberObjectToSchema))
}

func opOrganizationMemberImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid ID specified, should be in the format <organization_id>/<member_id>: '%s'", d.Id())
	}
	d.SetId(split[1])

	err := d.Set(schema_definition.AttributeOrganizationID, split[0])
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
