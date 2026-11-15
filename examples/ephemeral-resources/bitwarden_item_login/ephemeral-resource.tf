ephemeral "bitwarden_item_login" "example" {
  id = "37a66d6a-96c1-4f04-9a3c-b1fc0135669e"
}

# Alternatively, search for a uniquely named login item.
ephemeral "bitwarden_item_login" "by_search" {
  search = "VPN/Credentials"
}

# Pass ephemeral.bitwarden_item_login.example.password to a provider
# configuration or a write-only resource argument. Ordinary resource
# arguments and root outputs cannot accept ephemeral values. Prefer this
# over data.bitwarden_item_login when credentials should not be stored in state.
#
# Write-only arguments require Terraform 1.11+ and explicit
# support in the receiving provider. Use that resource's documented version
# or revision argument to trigger updates when the login changes.
