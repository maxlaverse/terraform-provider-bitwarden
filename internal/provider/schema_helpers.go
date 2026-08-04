package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

// clientsFromProviderData extracts ProviderClients from Framework provider data.
// A nil providerData is expected during validation/planning before Configure runs.
func clientsFromProviderData(providerData any, diags *diag.Diagnostics) (*ProviderClients, bool) {
	if providerData == nil {
		return nil, false
	}
	clients, ok := providerData.(*ProviderClients)
	if !ok {
		diags.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *ProviderClients, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil, false
	}
	return clients, true
}

// requirePasswordManager returns the Password Manager client from provider meta,
// or records a diagnostic and returns false.
func requirePasswordManager(clients *ProviderClients, diags *diag.Diagnostics) (bitwarden.PasswordManager, bool) {
	bwClient, err := clients.RequirePasswordManager()
	if err != nil {
		diags.AddError("Provider not configured for Password Manager", err.Error())
		return nil, false
	}
	return bwClient, true
}

// requireSecretsManager returns the Secrets Manager client from provider meta,
// or records a diagnostic and returns false.
func requireSecretsManager(clients *ProviderClients, diags *diag.Diagnostics) (bitwarden.SecretsManager, bool) {
	bwsClient, err := clients.RequireSecretsManager()
	if err != nil {
		diags.AddError("Provider not configured for Secrets Manager", err.Error())
		return nil, false
	}
	return bwsClient, true
}

// addErr reports a client/API error using the error text as the diagnostic
// summary, matching the historical SDKv2 diag.FromErr formatting that
// acceptance tests assert against (e.g. "Error: object not found").
func addErr(diags *diag.Diagnostics, err error) {
	diags.AddError(err.Error(), "")
}

// mapStr converts a map value from MapData into a Framework string attribute.
// Non-string values become null.
func mapStr(v interface{}) types.String {
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringNull()
}

type passwordManagerOperation func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) sdkdiag.Diagnostics
type secretsManagerOperation func(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) sdkdiag.Diagnostics

// withPasswordManager wraps an SDKv2 resource operation with a Password Manager client
// from provider meta. Kept until those resources migrate to Framework.
func withPasswordManager(resourceAction passwordManagerOperation) func(ctx context.Context, d *schema.ResourceData, meta interface{}) sdkdiag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) sdkdiag.Diagnostics {
		clients, ok := meta.(*ProviderClients)
		if !ok {
			return sdkdiag.FromErr(errPasswordManagerRequired)
		}
		bwClient, err := clients.RequirePasswordManager()
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		return resourceAction(ctx, d, bwClient)
	}
}

// withSecretsManager wraps an SDKv2 resource operation with a Secrets Manager client
// from provider meta. Kept until those resources migrate to Framework.
func withSecretsManager(resourceAction secretsManagerOperation) func(ctx context.Context, d *schema.ResourceData, meta interface{}) sdkdiag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) sdkdiag.Diagnostics {
		clients, ok := meta.(*ProviderClients)
		if !ok {
			return sdkdiag.FromErr(errSecretsManagerRequired)
		}
		bwsClient, err := clients.RequireSecretsManager()
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		return resourceAction(ctx, d, bwsClient)
	}
}
