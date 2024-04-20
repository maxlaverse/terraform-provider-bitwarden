data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

data "bitwarden_org_collection" "engineering" {
  search = "Engineering"
}

data "bitwarden_folder" "databases" {
  search = "Databases"
}

resource "bitwarden_item_secure_note" "example" {
  name            = "Example"
  notes           = <<EOT
[global]
secret = "<something sensitive>"
EOT
  folder_id       = data.bitwarden_folder.databases.id
  organization_id = data.bitwarden_organization.terraform.id
  collection_ids  = [data.bitwarden_org_collection.engineering.id]

  field {
    name = "category"
    text = "SystemA"
  }
}
