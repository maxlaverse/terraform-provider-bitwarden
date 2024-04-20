
data "bitwarden_organization" "terraform" {
  search = "Terraform"
}
data "bitwarden_org_collection" "engineering" {
  search          = "Engineering"
  organization_id = data.bitwarden_organization.terraform.id
}


# Example of usage of the data source:
resource "bitwarden_item_login" "administrative_user" {
  name     = "Service Administrator"
  username = "admin"
  password = "<sensitive>"

  organization_id = data.bitwarden_organization.terraform.id
  collection_ids  = [data.bitwarden_org_collection.terraform.id]
}
