package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opUserCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwClient.CreateUser, transformation.UserToObject, transformation.UserObjectToSchema))
}

func opUserRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		userEmail := d.Get(schema_definition.AttributeEmail).(string)
		obj, err := bwClient.FindUser(ctx, bitwarden.WithSearch(userEmail))
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.FromErr(transformation.UserObjectToSchema(ctx, obj, d))
	}
	return diag.FromErr(applyOperation(ctx, d, bwClient.GetUser, transformation.UserToObject, transformation.UserObjectToSchema))
}

func opUserReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwClient.GetUser, transformation.UserToObject, transformation.UserObjectToSchema))
}

func opUserDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwClient.DeleteUser), transformation.UserToObject, transformation.UserObjectToSchema))
}

func opUserImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	return []*schema.ResourceData{d}, nil
}
