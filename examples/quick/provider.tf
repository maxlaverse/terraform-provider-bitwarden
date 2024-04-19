terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.7.0"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email = "terraform@example.com"
}

# Create a Bitwarden Login Resource
resource "bitwarden_item_login" "example" {
  name     = "Example"
  username = "service-account"
  password = "<sensitive>"
}

# Use Bitwarden Resource
data "bitwarden_item_login" "example" {
  search     = "Example"
}