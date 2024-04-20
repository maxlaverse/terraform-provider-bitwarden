terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.8.0"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email = "terraform@example.com"
}

# Create a Bitwarden Login item
resource "bitwarden_item_login" "example" {
  name     = "Example"
  username = "service-account"
  password = "<sensitive>"
}

# or use an existing Bitwarden resource
data "bitwarden_item_login" "example" {
  search = "Example"
}