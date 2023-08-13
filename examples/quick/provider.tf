terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.6.1"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email           = "terraform@example.com"
}

# Create a Bitwarden Login Resource
resource "bitwarden_item_login" "example" {
  name            = "Example"
  username        = "service-account"
  password        = "<sensitive>"
}