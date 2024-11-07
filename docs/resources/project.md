---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bitwarden_project Resource - terraform-provider-bitwarden"
subcategory: ""
description: |-
  Manages a Project.
---

# bitwarden_project (Resource)

Manages a Project.

## Example Usage

```terraform
resource "bitwarden_project" "example" {
  name        = "Example Project"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name.

### Optional

- `id` (String) Identifier.
- `organization_id` (String) Identifier of the organization.

## Import

Import is supported using the following syntax:

```shell
$ terraform import bitwarden_project.example <project_id>
```