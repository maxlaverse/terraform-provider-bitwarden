provider "bitwarden" {
  server          = "https://vault.bitwarden.com"
  email           = "test@laverse.net"
  master_password = var.bw_password
  client_id       = var.bw_client_id
  client_secret   = var.bw_client_secret
}
