# Provider configuration
terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.5.0"
    }
  }
}

provider "bitwarden" {
  master_password = var.bw_password
  client_id       = var.bw_client_id
  client_secret   = var.bw_client_secret
  email           = "test@laverse.net"
  server          = "https://vault.bitwarden.com"
}