resource "bitwarden_item_login" "vpn_credentials" {
  name     = "VPN Admin Access"
  username = "admin"
}

resource "bitwarden_attachment" "vpn_config_from_content" {
  // NOTE: Only works when the experimental embedded client support is enabled
  file_name = "vpn-config.txt"
  content = jsonencode({
    domain : "laverse.net",
    persistence : {
      enabled : true,
    }
  })

  item_id = bitwarden_item_login.vpn_credentials.id
}

resource "bitwarden_attachment" "vpn_config_from_file" {
  file    = "vpn-config.txt"
  item_id = bitwarden_item_login.vpn_credentials.id
}
