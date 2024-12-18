package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type secretOperationFunc func(ctx context.Context, secret models.Secret) (*models.Secret, error)

func opSecretCreate() secretsManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
		return diag.FromErr(secretOperation(ctx, d, bwsClient.CreateSecret))
	}
}

func opSecretDelete(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(secretOperation(ctx, d, func(ctx context.Context, secret models.Secret) (*models.Secret, error) {
		return nil, bwsClient.DeleteSecret(ctx, secret)
	}))
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

	return diag.FromErr(secretOperation(ctx, d, func(ctx context.Context, secretReq models.Secret) (*models.Secret, error) {
		secret, err := bwsClient.GetSecret(ctx, secretReq)
		if secret != nil {
			if secret.ID != secretReq.ID {
				return nil, errors.New("returned secret ID does not match requested secret ID")
			}
		}

		return secret, err
	}))
}

func opSecretReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	err := secretOperation(ctx, d, func(ctx context.Context, secret models.Secret) (*models.Secret, error) {
		return bwsClient.GetSecret(ctx, secret)
	})

	if errors.Is(err, models.ErrObjectNotFound) {
		d.SetId("")
		tflog.Warn(ctx, "Secret not found, removing from state")
		return diag.Diagnostics{}
	}

	return diag.FromErr(err)
}

func opSecretUpdate(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(secretOperation(ctx, d, bwsClient.EditSecret))
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

	return transformation.SecretDataFromStruct(ctx, d, secret)
}

func secretOperation(ctx context.Context, d *schema.ResourceData, operation secretOperationFunc) error {
	secret, err := operation(ctx, transformation.SecretStructFromData(ctx, d))
	if err != nil {
		return err
	}

	return transformation.SecretDataFromStruct(ctx, d, secret)
}
