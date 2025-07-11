---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden_item_login Resource - terraform-provider-bitwarden"
subcategory: ""
description: |-
  Manages a login item.
---

# bitwarden_item_login (Resource)

Manages a login item.

## Example Usage

```terraform
data "bitwarden_organization" "terraform" {
  search = "Terraform"
}

data "bitwarden_org_collection" "engineering" {
  search = "Engineering"
}

data "bitwarden_folder" "databases" {
  search = "Databases"
}

resource "bitwarden_item_login" "administrative-user" {
  name     = "Service Administrator"
  username = "admin"
  password = "<sensitive>"

  folder_id       = data.bitwarden_folder.databases.id
  organization_id = bitwarden_organization.terraform.id
  collection_ids  = [bitwarden_org_collection.engineering.id]

  field {
    name = "category"
    text = "SystemA"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name.

### Optional

- `collection_ids` (List of String) Identifier of the collections the item belongs to.
- `favorite` (Boolean) Mark as a Favorite to have item appear at the top of your Vault in the UI.
- `field` (Block List) Extra fields. (see [below for nested schema](#nestedblock--field))
- `folder_id` (String) Identifier of the folder.
- `id` (String) Identifier.
- `notes` (String, Sensitive) Notes.
- `organization_id` (String) Identifier of the organization.
- `password` (String, Sensitive) Login password.
- `reprompt` (Boolean) Require master password 're-prompt' when displaying secret in the UI.
- `totp` (String, Sensitive) Verification code.
- `uri` (Block List) URI. (see [below for nested schema](#nestedblock--uri))
- `username` (String, Sensitive) Login username.

### Read-Only

- `attachments` (List of Object) List of item attachments. (see [below for nested schema](#nestedatt--attachments))
- `creation_date` (String) Date the item was created.
- `deleted_date` (String) Date the item was deleted.
- `revision_date` (String) Last time the item was updated.

<a id="nestedblock--field"></a>
### Nested Schema for `field`

Required:

- `name` (String) Name of the field.

Optional:

- `boolean` (Boolean) Value of a boolean field.
- `hidden` (String) Value of a hidden text field.
- `linked` (String) Value of a linked field.
- `text` (String) Value of a text field.


<a id="nestedblock--uri"></a>
### Nested Schema for `uri`

Required:

- `value` (String) URI Value

Optional:

- `match` (String) URI Match


<a id="nestedatt--attachments"></a>
### Nested Schema for `attachments`

Read-Only:

- `file_name` (String)
- `id` (String)
- `size` (String)
- `size_name` (String)
- `url` (String)

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
$ terraform import bitwarden_item_login.example <login_item_id>
```
