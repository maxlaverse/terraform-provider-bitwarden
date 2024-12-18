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
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opOrganizationCollectionCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(applyOperation(ctx, d, bwClient.CreateObject, transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}

func opOrganizationCollectionDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwClient.DeleteObject), transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}

func opOrganizationCollectionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func opOrganizationCollectionRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		return diag.FromErr(searchOperation(ctx, d, bwClient.ListObjects, transformation.ObjectDataFromStruct))
	}

	return diag.FromErr(applyOperation(ctx, d, bwClient.GetObject, transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}

func opOrganizationCollectionReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwClient.GetObject, transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}

func opOrganizationCollectionUpdate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(applyOperation(ctx, d, bwClient.EditObject, transformation.ObjectStructFromData, transformation.ObjectDataFromStruct))
}
