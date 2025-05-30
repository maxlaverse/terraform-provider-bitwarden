---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden_item_ssh_key Data Source - terraform-provider-bitwarden"
subcategory: ""
description: |-
  Use this data source to get information on an existing SSH key item.
---

# bitwarden_item_ssh_key (Data Source)

Use this data source to get information on an existing SSH key item.

## Example Usage

```terraform
data "bitwarden_item_ssh_key" "git_ssh_key" {
  search = "Git/SSH Key"
}

resource "local_sensitive_file" "id_rsa" {
  filename        = "id_rsa"
  file_permission = "600"
  content         = data.bitwarden_item_ssh_key.id_rsa.private_key
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filter_collection_id` (String) Filter search results by collection ID.
- `filter_folder_id` (String) Filter search results by folder ID.
- `filter_organization_id` (String) Filter search results by organization ID.
- `id` (String) Identifier.
- `search` (String) Search items matching the search string.

### Read-Only

- `collection_ids` (List of String) Identifier of the collections the item belongs to.
- `creation_date` (String) Date the item was created.
- `deleted_date` (String) Date the item was deleted.
- `field` (List of Object, Sensitive) Extra fields. (see [below for nested schema](#nestedatt--field))
- `folder_id` (String) Identifier of the folder.
- `key_fingerprint` (String, Sensitive) Key fingerprint.
- `name` (String) Name.
- `notes` (String, Sensitive) Notes.
- `organization_id` (String) Identifier of the organization.
- `private_key` (String, Sensitive) Private key.
- `public_key` (String, Sensitive) Public key.
- `reprompt` (Boolean) Require master password 're-prompt' when displaying secret in the UI.
- `revision_date` (String) Last time the item was updated.

<a id="nestedatt--field"></a>
### Nested Schema for `field`

Read-Only:

- `boolean` (Boolean)
- `hidden` (String)
- `linked` (String)
- `name` (String)
- `text` (String)
