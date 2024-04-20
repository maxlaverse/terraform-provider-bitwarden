data "bitwarden_folder" "terraform" {
  search = "Terraform"
}

# Example of usage of the data source:
resource "bitwarden_item_login" "administrative_user" {
  name     = "Service Administrator"
  username = "admin"

  folder_id = data.bitwarden_folder.terraform.id
}
