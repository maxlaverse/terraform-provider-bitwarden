package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opOrganizationRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))

	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrganization)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		return diag.FromErr(searchOperation(ctx, d, bwClient.ListObjects, transformation.BaseObjectToSchema))
	}

	return diag.FromErr(applyOperation(ctx, d, bwClient.GetObject, transformation.BaseSchemaToObject, transformation.BaseObjectToSchema))
}
