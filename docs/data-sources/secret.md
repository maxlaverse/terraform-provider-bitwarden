---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden_secret Data Source - terraform-provider-bitwarden"
subcategory: ""
description: |-
  Use this data source to get information on an existing secret.
---

# bitwarden_secret (Data Source)

Use this data source to get information on an existing secret.

## Example Usage

```terraform
data "bitwarden_secret" "example" {
  id = "37a66d6a-96c1-4f04-9a3c-b1fc0135669e"
}

resource "kubernetes_secret" "vpn_credentials" {
  metadata {
    name = "vpn-key"
  }

  data = {
    "PASSWORD" = data.bitwarden_secret.value
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) Identifier.
- `key` (String) Name.
- `organization_id` (String) Identifier of the organization.

### Read-Only

- `note` (String) Note.
- `project_id` (String) Identifier of the project.
- `value` (String) Value.