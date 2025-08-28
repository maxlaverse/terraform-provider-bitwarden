terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.16.0"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email = "terraform@example.com"

  # By default, the provider uses Bitwarden CLIs to interact with the remote
  # Vaults. You can also decide to a client embedded in the provider instead,
  # which removes the need for locally installed binaries.
  #
  # Learn more about the implications by reading the "Client Implementation"
  # section below.
  #
  # experimental {
  #   embedded_client = true
  # }
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
