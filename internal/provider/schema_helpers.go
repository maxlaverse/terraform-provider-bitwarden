package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

type passwordManagerOperation func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics
type secretsManagerOperation func(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics

func withPasswordManager(resourceAction passwordManagerOperation) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		bwClient, ok := meta.(bitwarden.PasswordManager)
		if !ok {
			return diag.FromErr(errors.New("provider was not configured with Password Manager credentials"))
		}
		return resourceAction(ctx, d, bwClient)
	}
}

func withSecretsManager(resourceAction secretsManagerOperation) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		bwsClient, ok := meta.(bitwarden.SecretsManager)
		if !ok {
			return diag.FromErr(errors.New("provider was not configured with Secrets Manager credentials"))
		}
		return resourceAction(ctx, d, bwsClient)
	}
}
