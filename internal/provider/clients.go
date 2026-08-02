package provider

import (
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

// ProviderClients holds the Bitwarden clients created during Configure.
// Exactly one of PasswordManager or SecretsManager is typically set, depending
// on the credentials supplied to the provider.
type ProviderClients struct {
	PasswordManager bitwarden.PasswordManager
	SecretsManager  bitwarden.SecretsManager
}

func (c *ProviderClients) RequirePasswordManager() (bitwarden.PasswordManager, error) {
	if c == nil || c.PasswordManager == nil {
		return nil, errPasswordManagerRequired
	}
	return c.PasswordManager, nil
}

func (c *ProviderClients) RequireSecretsManager() (bitwarden.SecretsManager, error) {
	if c == nil || c.SecretsManager == nil {
		return nil, errSecretsManagerRequired
	}
	return c.SecretsManager, nil
}
