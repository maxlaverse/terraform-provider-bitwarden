package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

type passwordManagerOperation func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics
type secretsManagerOperation func(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics

func withPasswordManager(resourceAction passwordManagerOperation) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		clients, ok := meta.(*ProviderClients)
		if !ok {
			return diag.FromErr(errPasswordManagerRequired)
		}
		bwClient, err := clients.RequirePasswordManager()
		if err != nil {
			return diag.FromErr(err)
		}
		return resourceAction(ctx, d, bwClient)
	}
}

func withSecretsManager(resourceAction secretsManagerOperation) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		clients, ok := meta.(*ProviderClients)
		if !ok {
			return diag.FromErr(errSecretsManagerRequired)
		}
		bwsClient, err := clients.RequireSecretsManager()
		if err != nil {
			return diag.FromErr(err)
		}
		return resourceAction(ctx, d, bwsClient)
	}
}
