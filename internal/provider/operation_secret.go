package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opSecretCreate(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwsClient.CreateSecret, transformation.SecretSchemaToObject, transformation.SecretObjectToSchema))
}

func opSecretDelete(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwsClient.DeleteSecret), transformation.SecretSchemaToObject, transformation.SecretObjectToSchema))
}

func opSecretImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	return []*schema.ResourceData{d}, nil
}

func opSecretRead(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		return diag.FromErr(secretSearch(ctx, d, bwsClient))
	}

	return diag.FromErr(applyOperation(ctx, d, bwsClient.GetSecret, transformation.SecretSchemaToObject, transformation.SecretObjectToSchema))
}

func opSecretReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwsClient.GetSecret, transformation.SecretSchemaToObject, transformation.SecretObjectToSchema))
}

func opSecretUpdate(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwsClient.EditSecret, transformation.SecretSchemaToObject, transformation.SecretObjectToSchema))
}

func secretSearch(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) error {
	secretKey, ok := d.GetOk(schema_definition.AttributeKey)
	if !ok {
		return fmt.Errorf("BUG: secret key not set in the resource data")
	}

	secret, err := bwsClient.GetSecretByKey(ctx, secretKey.(string))
	if err != nil {
		return err
	}

	return transformation.SecretObjectToSchema(ctx, secret, d)
}
