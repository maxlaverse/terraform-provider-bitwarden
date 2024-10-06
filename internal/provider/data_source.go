package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func resourceReadDataSourceItem(attrObject models.ObjectType, attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		err := d.Set(attributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}
		return resourceReadDataSourceObject(attrObject)(ctx, d, bwClient)
	}
}

func resourceReadDataSourceObject(objType models.ObjectType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		d.SetId(d.Get(attributeID).(string))
		err := d.Set(attributeObject, objType)
		if err != nil {
			return diag.FromErr(err)
		}
		return objectRead(ctx, d, bwClient)
	}
}

func resourceReadDataSourceSecret(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	d.SetId(d.Get(attributeID).(string))
	return secretRead(ctx, d, bwsClient)
}

func resourceReadDataSourceProject(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	d.SetId(d.Get(attributeID).(string))
	return projectRead(ctx, d, bwsClient)
}
