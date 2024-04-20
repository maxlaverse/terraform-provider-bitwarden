data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

# Organization 
resource "bitwarden_folder" "databases" {
  organization_id = data.bitwarden_organization.terraform.id
  name            = "Databases"
}

resource "bitwarden_item_login" "vpn_credentials" {
  organization_id = data.bitwarden_organization.terraform.id
  folder_id       = bitwarden_folder.databases.id

  name     = "VPN access for Databases"
  username = "example"
  password = random_password.example.result
}

data "bitwarden_item_secure_note" "vpn_preshared_key" {
  search = "VPN/Pre-sharedSecret"

  filter_organization_id = data.bitwarden_organization.terraform.id
  filter_folder_id       = bitwarden_folder.databases.id
}


resource "kubernetes_secret" "preshared_key" {
  metadata {
    name = "vpn-preshared-key"
  }

  data = {
    "private.key" = data.bitwarden_item_secure_note.vpn_preshared_key.notes
  }
}
