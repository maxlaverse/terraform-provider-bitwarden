# Requires Terraform 1.10+ or OpenTofu 1.11+ and an access_token.
# "cli" is also supported; it lists all accessible secrets and filters by ID.
provider "bitwarden" {
  client_implementation = "embedded"
}

ephemeral "bitwarden_secrets" "example" {
  ids = [
    "37a66d6a-96c1-4f04-9a3c-b1fc0135669e",
    "e8913312-c3f1-4d24-a411-11a007f487c9",
  ]
}

# Use ephemeral.bitwarden_secrets.example.values["37a66d6a-96c1-4f04-9a3c-b1fc0135669e"]
# in an ephemeral context, or pass the entire values map to a write-only argument.
# Write-only arguments require Terraform/OpenTofu 1.11+ and receiving provider
# support. Use that resource's version/revision argument to trigger rotation.
