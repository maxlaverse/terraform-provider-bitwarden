data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

# Example of usage of the data source:
resource "bitwarden_item_login" "administrative_user" {
  name     = "Service Administrator"
  username = "admin"
  password = "<sensitive>"

  organization_id = data.bitwarden_organization.terraform.id
}
