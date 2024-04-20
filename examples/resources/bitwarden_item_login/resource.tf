data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

data "bitwarden_org_collection" "engineering" {
  search = "Engineering"
}

data "bitwarden_folder" "databases" {
  search = "Databases"
}

resource "bitwarden_item_login" "administrative-user" {
  name     = "Service Administrator"
  username = "admin"
  password = "<sensitive>"

  folder_id       = data.bitwarden_folder.databases.id
  organization_id = bitwarden_organization.terraform.id
  collection_ids  = [bitwarden_org_collection.engineering.id]

  field {
    name = "category"
    text = "SystemA"
  }
}
