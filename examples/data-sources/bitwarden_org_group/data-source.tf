
data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

data "bitwarden_org_group" "engineers" {
  organization_id = data.bitwarden_organization.terraform.id
  filter_name     = "Engineers"
}
