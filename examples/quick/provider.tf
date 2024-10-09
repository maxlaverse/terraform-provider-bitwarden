terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.10.0"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email = "terraform@example.com"

  # If you have the opportunity, you can try out the embedded client which removes the need
  # for a locally installed Bitwarden CLI. Please note that this feature is still considered
  # as experimental, might not work as expected, and is not recommended for production use.
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
