data "bitwarden_item_login" "vpn_credentials" {
  search = "VPN/Credentials"
}


# Example of usage of the data source:
resource "kubernetes_secret" "vpn_credentials" {
  metadata {
    name = "vpn-credentials"
  }

  data = {
    "USERNAME" = data.bitwarden_item_secure_note.vpn_credentials.username
    "PASSWORD" = data.bitwarden_item_secure_note.vpn_credentials.password
  }
}
