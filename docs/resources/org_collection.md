---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden_org_collection Resource - terraform-provider-bitwarden"
subcategory: ""
description: |-
  Manages an organization collection.
---

# bitwarden_org_collection (Resource)

Manages an organization collection.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name.
- `organization_id` (String) Identifier of the organization.

### Optional

- `id` (String) Identifier.
- `member` (Block Set) [Experimental] Member (Users) of a collection. (see [below for nested schema](#nestedblock--member))
- `member_group` (Block Set) [Experimental] Member Groups of a collection. (see [below for nested schema](#nestedblock--member_group))

<a id="nestedblock--member"></a>
### Nested Schema for `member`

Required:

- `id` (String) [Experimental] Unique Identifier (UUID) of the user or group member.

Optional:

- `hide_passwords` (Boolean) [Experimental] Hide passwords.
- `manage` (Boolean) [Experimental] Can manage the collection.
- `read_only` (Boolean) [Experimental] Read/Write permissions.


<a id="nestedblock--member_group"></a>
### Nested Schema for `member_group`

Required:

- `id` (String) [Experimental] Unique Identifier (UUID) of the user or group member.

Optional:

- `hide_passwords` (Boolean) [Experimental] Hide passwords.
- `manage` (Boolean) [Experimental] Can manage the collection.
- `read_only` (Boolean) [Experimental] Read/Write permissions.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import bitwarden_org_collection.example <organization_id>/<collection_id>
```
