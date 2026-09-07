# Requires Terraform 1.10+ or OpenTofu 1.11+.
# Uses the access_token and client_implementation from the Bitwarden provider.
ephemeral "bitwarden_secret" "example" {
  id = "37a66d6a-96c1-4f04-9a3c-b1fc0135669e"
}

# Alternatively, look up a unique key accessible to the access token.
ephemeral "bitwarden_secret" "by_key" {
  key = "DATABASE_PASSWORD"
}

# Pass ephemeral.bitwarden_secret.example.value to a provider configuration
# or a write-only resource argument. Ordinary resource arguments and root
# outputs cannot accept ephemeral values.
#
# Write-only arguments require Terraform 1.11+ or OpenTofu 1.11+ and explicit
# support in the receiving provider. Use that resource's documented version
# or revision argument to trigger updates when the secret changes.
