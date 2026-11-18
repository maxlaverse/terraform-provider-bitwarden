ephemeral "bitwarden_attachment" "example" {
  id      = "4d6a41364d6a4dea8ddb1a"
  item_id = "37a66d6a-96c1-4f04-9a3c-b1fc0135669e"
}

# Pass ephemeral.bitwarden_attachment.example.content to a provider
# configuration or a write-only resource argument. Ordinary resource
# arguments and root outputs cannot accept ephemeral values. Prefer this
# over data.bitwarden_attachment when the content should not be stored in state.
#
# Write-only arguments require Terraform 1.11+ and explicit
# support in the receiving provider. Use that resource's documented version
# or revision argument to trigger updates when the attachment changes.
