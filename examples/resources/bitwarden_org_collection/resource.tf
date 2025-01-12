data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

resource "bitwarden_org_collection" "infrastructure" {
  name            = "Infrastructure Passwords"
  organization_id = data.bitwarden_organization.terraform.id
}

resource "bitwarden_org_collection" "generated" {
  name            = "Generated Passwords"
  organization_id = data.bitwarden_organization.terraform.id

  member {
    email = "devops@example.com"
  }
}
