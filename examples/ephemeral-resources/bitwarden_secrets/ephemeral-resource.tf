# Requires Terraform 1.10+ or OpenTofu 1.11+ and an access_token.
provider "bitwarden" {
  client_implementation = "embedded"
}

ephemeral "bitwarden_secrets" "example" {
  ids = [
    "37a66d6a-96c1-4f04-9a3c-b1fc0135669e",
    "e8913312-c3f1-4d24-a411-11a007f487c9",
  ]
}

# Use ephemeral.bitwarden_secrets.example.secrets["37a66d6a-96c1-4f04-9a3c-b1fc0135669e"].value
# in a provider configuration or a write-only resource argument.
# Write-only arguments require Terraform/OpenTofu 1.11+ and receiving provider
# support. Use that resource's documented version or revision argument to
# trigger updates when the secret changes.
