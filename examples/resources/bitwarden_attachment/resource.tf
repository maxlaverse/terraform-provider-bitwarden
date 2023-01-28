resource "bitwarden_item_login" "vpn_credentials" {
  name     = "VPN Admin Access"
  username = "admin"
}

resource "bitwarden_attachment" "vpn_config" {
  file    = "./vpn-config.txt"
  item_id = bitwarden_item_login.vpn_credentials.id
}