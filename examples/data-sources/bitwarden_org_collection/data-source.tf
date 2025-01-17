
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

# Example of usage with ACLs:
locals {
  emails =[
    "regular-user-1@example.com",
    "regular-user-2@example.com",
  ]
}

data "bitwarden_org_member" "regular_users" {
  organization_id = data.bitwarden_organization.terraform.id
  count          = length(local.emails)
  email          = local.emails[count.index]
}


resource "bitwarden_org_collection" "my_collection" {
  organization_id = data.bitwarden_organization.terraform.id
  name            = "my-collection"


  dynamic "member" {
    for_each = data.bitwarden_org_member.regular_users
    content {
        id = member.value.id
        read_only = true
    }
  }

  member {
    id = data.bitwarden_org_member.john.id
  }
}
