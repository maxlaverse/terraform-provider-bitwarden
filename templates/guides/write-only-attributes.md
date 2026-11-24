---
page_title: "Working with Secrets"
description: |-
  Patterns for retrieving, generating, and reusing Bitwarden-managed secrets without ever storing their plaintext in Terraform or OpenTofu state.
---

# Working with Secrets

This guide walks through three common situations and the combination of ephemeral resources and write-only arguments that solves each one, without a plaintext secret ever landing in a `.tfstate` file.

-> **Note:** Ephemeral resources require Terraform 1.10+ or OpenTofu 1.12+. Write-only arguments (`*_wo`) require Terraform 1.11+ or OpenTofu 1.12+, and explicit support in the attribute's provider. See [Using Ephemeral Resources](ephemeral-resources.md) for the general mechanics, lookup arguments, and the full list of ephemeral resources this provider exposes.

## I want to retrieve a password safely

You already have a login item in Bitwarden and need its password somewhere in your configuration - a database provider's `password` argument, another resource's write-only argument, and so on - without it ever being written to state.

Use the `bitwarden_item_login` ephemeral resource and pass its `password` straight through:

```terraform
ephemeral "bitwarden_item_login" "db" {
  search = "Database/Postgres"
}

provider "postgresql" {
  host     = "db.example.com"
  username = ephemeral.bitwarden_item_login.db.username
  password = ephemeral.bitwarden_item_login.db.password
}
```

This is enforced by Terraform itself, not just a convention: an ephemeral value is tagged as such all the way through your configuration, and Terraform refuses to plan if you try to route it into a context that would persist it - an ordinary (non write-only) resource argument, or a root module `output` (only child module outputs can be marked `ephemeral`). You don't have to remember to keep it out of state; attempting to put it there is a plan-time error, not a possible mistake.

The practical effect: the password is fetched fresh from Bitwarden on every plan and apply, used immediately, and then discarded. It is never diffed, never cached, and never appears in `terraform show` or any state file.

## I want to generate and store a password safely

You don't have a password yet - you want Terraform to generate one and create the Bitwarden item for it, again without the plaintext ever sitting in state.

Combine an ephemeral generator (`hashicorp/random`'s `random_password` ephemeral resource) with `bitwarden_item_login`'s write-only `password_wo` and its `wo_version` gate:

```terraform
ephemeral "random_password" "db" {
  length = 24
}

resource "bitwarden_item_login" "db" {
  name        = "Database/Postgres"
  username    = "app"
  password_wo = ephemeral.random_password.db.result
  wo_version  = 1
}
```

Two things to note:

- `random_password`'s *ephemeral* resource (as opposed to its ordinary resource) never stores the generated value anywhere, so unlike `resource "random_password"`, there's no copy of the plaintext sitting in `random_password`'s own state either - the guarantee is end to end.
- Because nothing remembers the ephemeral value between runs, `ephemeral.random_password.db.result` is a *different* string on every single plan and apply. Without `wo_version`, that would mean every `terraform apply` rotates the vault password just from being re-run. `wo_version` is what makes this idempotent: `password_wo` is only re-applied to the vault item when `wo_version` actually changes since the last apply. Leave it alone, and repeated applies are a no-op for the password, no matter what the generator produces that run.

**To rotate the password**, bump `wo_version`:

```terraform
resource "bitwarden_item_login" "db" {
  name        = "Database/Postgres"
  username    = "app"
  password_wo = ephemeral.random_password.db.result
  wo_version  = 2 # was 1
}
```

This updates the existing item in place: same item ID, and only the password changes. Keep configuring the other attributes you care about (`notes`, `totp`, `folder_id`, `collection_ids`) in the resource, because an update sends the whole item to Bitwarden.

The same applies when you switch an existing item from `password` to `password_wo`: the new value is applied and the old secret leaves the state, even if `wo_version` stays the same. After `terraform import`, the state mirrors the vault (plaintext password included) and has no `wo_version`, so the first apply with `password_wo` always applies the new value. Set the `wo_version` you want and expect that one rotation.

Avoid reaching for `terraform taint` or `-replace` to force a new password instead. Both force a destroy-and-recreate of the item, which gives it a **new ID** in Bitwarden - anything outside Terraform that stored the old ID (scripts, other tools, shared links) stops working, and you've generated far more vault churn (and audit log noise) than the rotation needed. Bumping `wo_version` is the supported way to rotate; replacement is a last resort for when the item itself needs to be rebuilt, not a rotation mechanism.

## I want to generate a password, store it safely, and reuse it in my infra

You need the same password in more than one place - say, the database itself and an application's configuration - and you want every consumer to agree on the current value, including after a rotation or a replacement of the Bitwarden item.

The trap is reaching for the generator again at every consumer:

```terraform
ephemeral "random_password" "db" {
  length = 24
}

resource "bitwarden_item_login" "db" {
  name        = "Database/Postgres"
  password_wo = ephemeral.random_password.db.result
  wo_version  = 1
}

resource "postgresql_role" "app" {
  name                = "app"
  password_wo         = ephemeral.random_password.db.result # looks fine, but it isn't
  password_wo_version = 1
}
```

Within a single apply that happens to work, because Terraform opens an ephemeral resource instance once per phase (plan, then apply) and reuses the value for every reference to that same address within that phase. But on the *next* apply, if `wo_version` hasn't changed, `bitwarden_item_login.db` deliberately keeps the password it already has - while `ephemeral.random_password.db.result` generates a brand-new, different string anyway, since it has no memory of the previous run. Any resource reading straight from the generator now disagrees with what's actually set on the vault item. The mismatch is silent until something fails to authenticate.

The fix is to treat the Bitwarden item as the single source of truth: generate and store the password once (as in the previous section), then have every consumer read it back *from Bitwarden*, not from the original generator:

```terraform
ephemeral "random_password" "db" {
  length = 24
}

resource "bitwarden_item_login" "db" {
  name        = "Database/Postgres"
  username    = "app"
  password_wo = ephemeral.random_password.db.result
  wo_version  = 1
}

ephemeral "bitwarden_item_login" "db_creds" {
  id = bitwarden_item_login.db.id
}

resource "postgresql_role" "app" {
  name                = "app"
  password_wo         = ephemeral.bitwarden_item_login.db_creds.password
  password_wo_version = bitwarden_item_login.db.wo_version
}
```

Note that `password_wo_version` is not a separate number: it reads `wo_version` straight from the Bitwarden resource. This matters because Terraform cannot diff a write-only value, so the version is the *only* signal `postgresql_role` has that it should push a new password. If it were hardcoded or tracked in its own variable, bumping `wo_version` on `bitwarden_item_login` would rotate the password in Bitwarden while PostgreSQL silently kept the old one, until someone remembered to bump the second counter too. Referencing the attribute (or, if you prefer, a single `local` that feeds every `*_wo_version`) makes one bump rotate every copy in the same apply, and Terraform orders the writes correctly because the consumers depend on `bitwarden_item_login.db`.

Any other consumer works the same way, as long as the argument it hands the password to is itself write-only (or a provider configuration argument, as in the first section): an ordinary argument still can't accept an ephemeral value, even one read back from Bitwarden.

-> **Note:** This keeps consumers in sync for rotations done by bumping `wo_version`. Replacing the Bitwarden item (`-replace`, `taint`, or a forced replacement) also produces a new password, but leaves `wo_version` unchanged, so a version-gated consumer such as `postgresql_role` would not notice. Consumers that read the password on every apply without a version gate (a provider configuration argument, for example) are unaffected. Another reason to rotate by bumping `wo_version` rather than by replacing the item.

The `bitwarden_item_login` resource is the only thing that ever *writes* the password, gated by `wo_version` as described above. Every consumer *reads* it back through the ephemeral resource, which always reflects whatever is currently stored in the vault. That's why both halves matter together: the resource exists to control when the password changes, and the ephemeral read-back exists so that anything depending on the current value gets it from the one place it's actually kept, rather than re-deriving its own guess. Consumers never hold their own copy of the password to drift, and the only thing to keep aligned is the version number, which is why it is referenced rather than repeated.
