package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceReadDataSourceItem(attrObject models.ObjectType, attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		err := d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}
		return resourceReadDataSourceObject(attrObject)(ctx, d, bwClient)
	}
}

func resourceReadDataSourceObject(objType models.ObjectType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		d.SetId(d.Get(schema_definition.AttributeID).(string))
		err := d.Set(schema_definition.AttributeObject, objType)
		if err != nil {
			return diag.FromErr(err)
		}
		return objectRead(ctx, d, bwClient)
	}
}

func resourceReadDataSourceSecret(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	return secretRead(ctx, d, bwsClient)
}

func resourceReadDataSourceProject(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	return projectRead(ctx, d, bwsClient)
}
