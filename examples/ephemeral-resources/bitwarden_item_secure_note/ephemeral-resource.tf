ephemeral "bitwarden_item_secure_note" "example" {
  id = "37a66d6a-96c1-4f04-9a3c-b1fc0135669e"
}

# Alternatively, search for a uniquely named secure note item.
ephemeral "bitwarden_item_secure_note" "by_search" {
  search = "VPN/Pre-sharedSecret"
}

# Pass ephemeral.bitwarden_item_secure_note.example.notes to a provider
# configuration or a write-only resource argument. Ordinary resource
# arguments and root outputs cannot accept ephemeral values. Prefer this
# over data.bitwarden_item_secure_note when the note should not be stored in state.
#
# Write-only arguments require Terraform 1.11+ and explicit
# support in the receiving provider. Use that resource's documented version
# or revision argument to trigger updates when the note changes.
