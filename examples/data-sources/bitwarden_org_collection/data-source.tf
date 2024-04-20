data "bitwarden_org_collection" "terraform" {
  search = "Terraform"
}

resource "bitwarden_item_login" "administrative_user" {
  name            = "Service Administrator"
  username        = "admin"
  password        = "<sensitive>"

  organization_id = "54421e78-95cb-40c4-a257-17231a7b6207"
  collection_ids  = [data.bitwarden_org_collection.terraform.id]
}
