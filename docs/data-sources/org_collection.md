---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden_org_collection Data Source - terraform-provider-bitwarden"
subcategory: ""
description: |-
  Use this data source to get information on an existing organization collection.
---

# bitwarden_org_collection (Data Source)

Use this data source to get information on an existing organization collection.

## Example Usage

```terraform
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
  emails = [
    "regular-user-1@example.com",
    "regular-user-2@example.com",
  ]
}

data "bitwarden_org_member" "regular_users" {
  organization_id = data.bitwarden_organization.terraform.id
  count           = length(local.emails)
  email           = local.emails[count.index]
}


resource "bitwarden_org_collection" "my_collection" {
  organization_id = data.bitwarden_organization.terraform.id
  name            = "my-collection"


  dynamic "member" {
    for_each = data.bitwarden_org_member.regular_users
    content {
      id        = member.value.id
      read_only = true
    }
  }

  member {
    id = data.bitwarden_org_member.john.id
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `organization_id` (String) Identifier of the organization.

### Optional

- `id` (String) Identifier.
- `search` (String) Search items matching the search string.

### Read-Only

- `member` (Set of Object) [Experimental] Member (Users) of a collection. (see [below for nested schema](#nestedatt--member))
- `member_group` (Set of Object) [Experimental] Member Groups of a collection. (see [below for nested schema](#nestedatt--member_group))
- `name` (String) Name.

<a id="nestedatt--member"></a>
### Nested Schema for `member`

Read-Only:

- `hide_passwords` (Boolean)
- `id` (String)
- `manage` (Boolean)
- `read_only` (Boolean)


<a id="nestedatt--member_group"></a>
### Nested Schema for `member_group`

Read-Only:

- `hide_passwords` (Boolean)
- `id` (String)
- `manage` (Boolean)
- `read_only` (Boolean)
