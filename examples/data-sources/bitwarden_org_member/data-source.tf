
data "bitwarden_organization" "terraform" {
  search = "Terraform"
}
data "bitwarden_org_member" "john" {
  email          = "john@example.com"
  organization_id = data.bitwarden_organization.terraform.id
}


# Example of usage of the data source:
# See org_collection
