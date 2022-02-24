variable "terraform_organization" {
  type        = string
  description = "Terraform Organization Identifier"
  default     = "ac901e49-5417-46e2-95fa-baf63186f751"
}

# Save sensitive Terraform generated data to Bitwarden
resource "bitwarden_folder" "terraform-bw-folder" {
  name = "Terraform Generated"
}

resource "bitwarden_item_login" "vpn-read-only-userpwd" {
  name      = "VPN Read Only User/Password Access"
  username  = "vpn-read-only"
  password  = some_other_plugin.user-read-only.secret
  folder_id = bitwarden_folder.terraform-bw-folder.id
}

resource "bitwarden_item_secure_note" "vpn-read-only-certs" {
  name            = "VPN Read Only Certificate Access"
  notes           = some_other_plugin.user-read-only.private_key
}

# Read sensitive information from Bitwarden
data "bitwarden_item_login" "mysql-root-credentials" {
  id = "ec4e447f-9aed-4203-b834-c8f3848828f7"
}

# Later to be accessed as
#   ... = data.bitwarden_item_login.mysql-root-credentials.username
#   ... = data.bitwarden_item_login.mysql-root-credentials.password

data "bitwarden_item_secure_note" "ssh-private-key" {
  id = "a9e19f26-1b8c-4568-bc09-191e2cf56ed6"
}

# ....