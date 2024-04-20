data "bitwarden_folder" "terraform" {
  search = "Terraform"
}

resource "bitwarden_item_login" "administrative_user" {
  name            = "Service Administrator"
  username        = "admin"
  password        = "<sensitive>"

  folder_id  = data.bitwarden_folder.terraform.id
}
