---
page_title: "Using Ephemeral Resources"
description: |-
  Read Bitwarden secrets and items without storing their values in Terraform or OpenTofu plan and state files.
---

# Using Ephemeral Resources

Ephemeral resources read values from Bitwarden for the current plan or apply, then discard them. Unlike data sources, those values are not written to plan or state files.

Use them when a credential should be passed into another provider configuration or a write-only resource argument, and should not persist in state.

-> **Note:** Ephemeral resources require Terraform 1.10 or later, or OpenTofu 1.12 or later. Write-only arguments on the *receiving* resource require Terraform 1.11+ (or equivalent OpenTofu support) and explicit support in that provider.

## Available resources

| Type | Product | Typical value to pass on |
|------|---------|--------------------------|
| `bitwarden_secret` | Secrets Manager | `value` |
| `bitwarden_item_login` | Password Manager | `username`, `password`, `totp` |
| `bitwarden_item_ssh_key` | Password Manager | `private_key` |
| `bitwarden_item_secure_note` | Password Manager | `notes` |

Each type has a matching data source. Prefer the ephemeral resource unless you need the value in an ordinary resource argument or a root output, which cannot accept ephemeral values.

Full schemas are on the Terraform Registry under [Ephemeral Resources](https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs). OpenTofu's registry does not list those pages yet; use this guide plus the generated files in `docs/ephemeral-resources/` in the provider repository.

## Example: provider configuration

```terraform
ephemeral "bitwarden_secret" "db" {
  key = "DATABASE_PASSWORD"
}

provider "postgresql" {
  host     = "db.example.com"
  username = "app"
  password = ephemeral.bitwarden_secret.db.value
}
```

Password Manager items work the same way:

```terraform
ephemeral "bitwarden_item_login" "vpn" {
  search = "VPN/Credentials"
}

provider "some_vpn" {
  username = ephemeral.bitwarden_item_login.vpn.username
  password = ephemeral.bitwarden_item_login.vpn.password
}
```

## Write-only arguments

A write-only argument (`*_wo`) on another resource can take an ephemeral value without storing it. The receiving resource usually also has a version or revision argument that you bump when the secret should be re-applied.

```terraform
ephemeral "bitwarden_secret" "db" {
  key = "DATABASE_PASSWORD"
}

resource "aws_secretsmanager_secret" "db" {
  name = "database-password"
}

resource "aws_secretsmanager_secret_version" "db" {
  secret_id                = aws_secretsmanager_secret.db.id
  secret_string_wo         = ephemeral.bitwarden_secret.db.value
  secret_string_wo_version = 1
}
```

This example uses the AWS provider's write-only `secret_string_wo` argument. Any provider that exposes a compatible write-only argument works the same way. See that resource's documentation for the version or revision argument to bump when the secret changes.

Lookup arguments (`id`, `key`, `search`, and filters) and the full attribute list for each ephemeral type are documented with the schema on the Terraform Registry.
