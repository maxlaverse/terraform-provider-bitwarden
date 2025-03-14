---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden_item_secure_note Data Source - terraform-provider-bitwarden"
subcategory: ""
description: |-
  Use this data source to get information on an existing secure note item.
---

# bitwarden_item_secure_note (Data Source)

Use this data source to get information on an existing secure note item.

## Example Usage

```terraform
data "bitwarden_item_secure_note" "vpn_preshared_key" {
  search = "VPN/Pre-sharedSecret"
}


# Example of usage of the data source:
resource "kubernetes_secret" "preshared_key" {
  metadata {
    name = "vpn-preshared-key"
  }

  data = {
    "private.key" = data.bitwarden_item_secure_note.vpn_preshared_key.notes
  }
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

- `attachments` (List of Object) List of item attachments. (see [below for nested schema](#nestedatt--attachments))
- `collection_ids` (List of String) Identifier of the collections the item belongs to.
- `creation_date` (String) Date the item was created.
- `deleted_date` (String) Date the item was deleted.
- `favorite` (Boolean) Mark as a Favorite to have item appear at the top of your Vault in the UI.
- `field` (List of Object, Sensitive) Extra fields. (see [below for nested schema](#nestedatt--field))
- `folder_id` (String) Identifier of the folder.
- `name` (String) Name.
- `notes` (String, Sensitive) Notes.
- `organization_id` (String) Identifier of the organization.
- `reprompt` (Boolean) Require master password 're-prompt' when displaying secret in the UI.
- `revision_date` (String) Last time the item was updated.

<a id="nestedatt--attachments"></a>
### Nested Schema for `attachments`

Read-Only:

- `file_name` (String)
- `id` (String)
- `size` (String)
- `size_name` (String)
- `url` (String)


<a id="nestedatt--field"></a>
### Nested Schema for `field`

Read-Only:

- `boolean` (Boolean)
- `hidden` (String)
- `linked` (String)
- `name` (String)
- `text` (String)
