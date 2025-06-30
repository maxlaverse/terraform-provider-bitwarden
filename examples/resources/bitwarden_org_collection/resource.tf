data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

data "bitwarden_org_member" "john" {
  email           = "john@example.com"
  organization_id = data.bitwarden_organization.terraform.id
}

data "bitwarden_org_group" "johns_group" {
  organization_id = data.bitwarden_organization.terraform.id
  filter_name     = "John's Group"
}

resource "bitwarden_org_collection" "infrastructure" {
  name            = "Infrastructure Passwords"
  organization_id = data.bitwarden_organization.terraform.id
}

resource "bitwarden_org_collection" "generated" {
  name            = "Generated Passwords"
  organization_id = data.bitwarden_organization.terraform.id

  member {
    id             = data.bitwarden_org_member.john.id
    hide_passwords = false
    read_only      = false
    manage         = true
  }
  member_group {
    id             = data.bitwarden_org_group.johns_group.id
    hide_passwords = false
    read_only      = false
    manage         = true
  }
}

