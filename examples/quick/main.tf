variable "terraform_organization" {
  type        = string
  description = "My Cloud Organization"
  default     = "ac901e49-5417-46e2-95fa-baf63186f751"
}

resource "bitwarden_folder" "cloud_credentials" {
  organization_id = var.terraform_organization
  name            = "My Cloud Credentials"
}

resource "bitwarden_item_login" "vpn_credentials" {
  organization_id = var.terraform_organization
  folder_id       = bitwarden_folder.cloud_credentials.id

  name     = "VPN Read Only User/Password Access"
  username = "vpn-user"
  password = random_password.vpn_password.result
}

resource "bitwarden_item_secure_note" "vpn_note" {
  organization_id = var.terraform_organization
  folder_id       = bitwarden_folder.cloud_credentials.id

  name  = "Notes on the preshared Secret"
  notes = "It's 1234"
}


# Read sensitive information from Bitwarden
# Using Login information
data "bitwarden_item_login" "mysql_credentials" {
  id = "ec4e447f-9aed-4203-b834-c8f3848828f7"
}

# Later to be accessed as
#   ... = data.bitwarden_item_login.mysql-root-credentials.username
#   ... = data.bitwarden_item_login.mysql-root-credentials.password
#
# or for fields:
# lookup(
#    zipmap(
#      data.bitwarden_item_login.mysql-root-credentials.field.*.name,
#      data.bitwarden_item_login.mysql-root-credentials.field.*
#    ),
#    "<name-of-the-field-your-looking-for>"
# )

# Using Attachments
data "bitwarden_attachment" "ssh_credentials" {
  id      = "4d6a41364d6a4dea8ddb1a"
  item_id = "59575167-4d36-5a58-466e-d9021926df8a"
}

resource "kubernetes_secret" "ssh" {
  metadata {
    name = "ssh"
  }

  data = {
    "private.key" = data.bitwarden_attachment.ssh_credentials.content
  }
}
# ....