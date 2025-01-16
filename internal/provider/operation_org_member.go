package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

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
