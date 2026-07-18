data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

resource "bitwarden_org_member" "john" {
  organization_id = data.bitwarden_organization.terraform.id
  email           = "john@example.com"
  role            = "user" # one of: owner, admin, user, manager
}
