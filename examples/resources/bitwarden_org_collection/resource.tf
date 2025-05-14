variable "BW_GROUP_ID" {
  type        = string
  description = "Static UUID for a Bitwarden Group"
}

data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

data "bitwarden_org_member" "john" {
  email           = "john@example.com"
  organization_id = data.bitwarden_organization.terraform.id
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
    id             = var.BW_GROUP_ID
    hide_passwords = false
    read_only      = false
    manage         = true
  }
}

