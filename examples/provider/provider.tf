provider "bitwarden" {
  master_password = var.bw_password
  email           = "test@laverse.net"
  server          = "https://vault.bitwarden.com"
}
