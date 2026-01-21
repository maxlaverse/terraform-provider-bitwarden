terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.17.1"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email = "terraform@example.com"

  # By default, the provider uses Bitwarden CLIs to interact with the remote
  # Vaults. You can also use the embedded client instead, which removes
  # the need for locally installed binaries.
  #
  # Learn more about the implications by reading the "Client Implementation"
  # section below.
  #
  # client_implementation = "embedded"
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
