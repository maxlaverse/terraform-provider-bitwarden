data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

resource "bitwarden_org_collection" "engineering" {
  name            = "Engineering"
  organization_id = data.bitwarden_organization.terraform.id
}
