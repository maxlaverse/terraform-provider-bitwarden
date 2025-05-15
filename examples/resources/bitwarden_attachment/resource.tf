# Create a Login to attach items to.
resource "bitwarden_item_login" "vpn_credentials" {
  name     = "VPN Admin Access"
  username = "admin"
}

# 1. Attach raw content to the login.
resource "bitwarden_attachment" "vpn_config_from_content" {
  // NOTE: Only works when the experimental embedded client support is enabled
  file_name = "vpn-config.txt"
  content = jsonencode({
    domain : "example.com",
    persistence : {
      enabled : true,
    }
  })
  item_id = bitwarden_item_login.vpn_credentials.id
}

# 2. Attach an existing file to the login.
resource "bitwarden_attachment" "vpn_config_from_file" {
  file    = "vpn-config.txt"
  item_id = bitwarden_item_login.vpn_credentials.id
}

# 3. Generate a dynamic file and attach it to the login
resource "tls_private_key" "rsa_key" {
  algorithm = "RSA"
  # rsa_bits  = 2048
  rsa_bits = 4096
}

resource "local_sensitive_file" "vpn_key_file" {
  filename        = "vpn_key.pem"
  file_permission = "600"
  content         = tls_private_key.rsa_key.private_key_pem
}

resource "bitwarden_attachment" "dynamic_key_from_disk" {
  file = local_sensitive_file.vpn_key_file.filename
  # (Optional) Hash specified ensures the file is reuploaded if it is recalculated.
  content_hash = local_sensitive_file.vpn_key_file.content_sha1
  item_id      = bitwarden_item_login.vpn_credentials.id
}

resource "bitwarden_attachment" "dynamic_key_from_content" {
  content = local_sensitive_file.vpn_key_file.content
  item_id = bitwarden_item_login.vpn_credentials.id
}