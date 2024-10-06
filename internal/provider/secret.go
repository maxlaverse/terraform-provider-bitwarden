package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

type secretOperationFunc func(ctx context.Context, secret models.Secret) (*models.Secret, error)

func resourceCreateSecret() secretsManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
		return secretCreate(ctx, d, bwsClient)
	}
}

func resourceReadSecretIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
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

func resourceUpdateSecret(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(secretOperation(ctx, d, bwsClient.EditSecret))
}

func resourceDeleteSecret(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(secretOperation(ctx, d, func(ctx context.Context, secret models.Secret) (*models.Secret, error) {
		return nil, bwsClient.DeleteSecret(ctx, secret)
	}))
}

func secretCreate(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(secretOperation(ctx, d, bwsClient.CreateSecret))
}

func secretRead(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
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

func secretOperation(ctx context.Context, d *schema.ResourceData, operation secretOperationFunc) error {
	secret, err := operation(ctx, secretStructFromData(ctx, d))
	if err != nil {
		return err
	}

	return secretDataFromStruct(ctx, d, secret)
}

func secretStructFromData(_ context.Context, d *schema.ResourceData) models.Secret {
	var secret models.Secret

	secret.ID = d.Id()
	if v, ok := d.Get(attributeKey).(string); ok {
		secret.Key = v
	}

	if v, ok := d.Get(attributeValue).(string); ok {
		secret.Value = v
	}

	if v, ok := d.Get(attributeNote).(string); ok {
		secret.Note = v
	}

	if v, ok := d.Get(attributeOrganizationID).(string); ok {
		secret.OrganizationID = v
	}

	if v, ok := d.Get(attributeProjectID).(string); ok {
		secret.ProjectID = v
	}

	return secret
}

func secretDataFromStruct(_ context.Context, d *schema.ResourceData, secret *models.Secret) error {
	if secret == nil {
		// Secret has been deleted
		return nil
	}

	d.SetId(secret.ID)

	err := d.Set(attributeKey, secret.Key)
	if err != nil {
		return err
	}

	err = d.Set(attributeValue, secret.Value)
	if err != nil {
		return err
	}

	err = d.Set(attributeNote, secret.Note)
	if err != nil {
		return err
	}

	err = d.Set(attributeOrganizationID, secret.OrganizationID)
	if err != nil {
		return err
	}

	err = d.Set(attributeProjectID, secret.ProjectID)
	if err != nil {
		return err
	}

	return nil
}

func resourceImportSecret(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	return []*schema.ResourceData{d}, nil
}
