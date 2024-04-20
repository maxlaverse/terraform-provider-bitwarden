---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden_folder Data Source - terraform-provider-bitwarden"
subcategory: ""
description: |-
  Use this data source to get information on an existing folder.
---

# bitwarden_folder (Data Source)

Use this data source to get information on an existing folder.

## Example Usage

```terraform
data "bitwarden_folder" "terraform" {
  search = "Terraform"
}

# Example of usage of the data source:
resource "bitwarden_item_login" "administrative_user" {
  name     = "Service Administrator"
  username = "admin"

  folder_id = data.bitwarden_folder.terraform.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filter_collection_id` (String) Filter search results by collection ID.
- `filter_organization_id` (String) Filter search results by organization ID.
- `id` (String) Identifier.
- `search` (String) Search items matching the search string.

### Read-Only

- `name` (String) Name.