data "bitwarden_item_secure_note" "vpn_preshared_key" {
  search = "VPN/Pre-sharedSecret"
}


# Example of usage of the data source:
resource "kubernetes_secret" "preshared_key" {
  metadata {
    name = "vpn-preshared-key"
  }

  data = {
    "private.key" = data.bitwarden_item_secure_note.vpn_preshared_key.notes
  }
}
