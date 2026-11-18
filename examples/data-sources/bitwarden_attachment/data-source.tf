# Values from this data source are stored in Terraform state. Use
# ephemeral.bitwarden_attachment when the content should not be persisted.
data "bitwarden_item_login" "ssh" {
  search = "VPN/Credentials"
}

data "bitwarden_attachment" "ssh_private_key" {
  id      = "4d6a41364d6a4dea8ddb1a"
  item_id = data.bitwarden_item_login.ssh.id
}

# Example of usage of the data source:
resource "kubernetes_secret" "ssh_keys" {
  metadata {
    name = "ssh-keys"
  }

  data = {
    "private.key" = data.bitwarden_attachment.ssh_private_key.content
  }
}
