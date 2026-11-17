---
page_title: "Authentication"
description: |-
  Authenticate the Bitwarden provider with Password Manager or Secrets Manager credentials.
---

# Authentication

Depending on the credentials you set, the provider talks to either [Password Manager] or [Secrets Manager].
To use both in the same workspace, configure [provider aliases](#using-password-manager-and-secrets-manager-together).

Argument names and environment variables are listed on the [provider overview](../index).

## Password Manager

Password Manager accepts these combinations:

* _[Personal API Key]_ (requires `master_password`, `client_id`, and `client_secret`). This is the recommended method. Set `email` as well when using the CLI client so the provider can tell which local Vault belongs to you.
* _Email and Password_ (requires `email` and `master_password`). Prefer a Personal API Key instead.
* User-provided _Session Key_ (requires `session_key`), which only works with a pre-downloaded Vault (see [Generating a Session Key](#generating-a-session-key)).

### Generating a Client ID and Secret

A Personal API Key lets you revoke access to your Vault without changing your master password.

1. Connect to your Vault on https://vault.bitwarden.com, or your self-hosted instance
2. Click on _Settings_ and then _My Account_
3. Scroll down to the _API Key_ section
4. Click on _View API Key_ (or maybe another label if it's the first time)
5. Save the API credentials somewhere safe

```terraform
provider "bitwarden" {
  email           = var.email
  master_password = var.master_password
  client_id       = var.client_id
  client_secret   = var.client_secret
}
```

### Generating a Session Key

If you don't want to use an API key, you can use a Session Key instead.
When doing so, it's your responsibility to:

* ensure the validity of the Session Key
* keep the Session Key safe
* revoke it when you don't need it anymore

You can generate a Session Key by running the following command in your Terraform workspace:

```console
BITWARDENCLI_APPDATA_DIR=.bitwarden bw login

# or if you use a custom vault path
BITWARDENCLI_APPDATA_DIR=<vault_path> bw login
```

A Session Key is bound to a local copy of a Vault. Set `BITWARDENCLI_APPDATA_DIR` (or `vault_path`) to the directory where that Vault is stored.

```terraform
provider "bitwarden" {
  email       = var.email
  session_key = var.session_key
}
```

## Secrets Manager

Secrets Manager only accepts [Access Tokens] (`access_token`).

To generate an Access Token:

1. Connect to your Vault on https://vault.bitwarden.com
2. Ensure the _Secrets Manager_ section is selected (bottom left)
3. Click on _Machine accounts_
4. Click on _New_
5. Click on your generated Machine Account
6. Select the _Access Tokens_ tab
7. Created a new Access Token and save it somewhere safe

```terraform
provider "bitwarden" {
  access_token = var.access_token
}
```

-> **Note:** By default, new machine accounts are not assigned to any projects. After creating a machine account, make sure it is added to the project(s) you want to access, with either "Can Read" or "Can Read & Write" permissions. If the machine account does not have access to the project, you will encounter an `Error: object not found` when attempting to use/create secrets from that project.

## Using Password Manager and Secrets Manager together

A single provider configuration authenticates to one product. Use [aliases][provider aliases] when a workspace needs both.

```terraform
provider "bitwarden" {
  alias           = "pm"
  email           = var.email
  master_password = var.master_password
  client_id       = var.client_id
  client_secret   = var.client_secret
}

provider "bitwarden" {
  alias        = "sm"
  access_token = var.access_token
}

data "bitwarden_item_login" "vpn" {
  provider = bitwarden.pm
  search   = "VPN"
}

data "bitwarden_secret" "api_key" {
  provider = bitwarden.sm
  key      = "API_KEY"
}
```

To read credentials without storing them in plan or state, use [ephemeral resources](ephemeral-resources) instead of data sources.

[Password Manager]: https://bitwarden.com/products/personal/
[Secrets Manager]: https://bitwarden.com/products/secrets-manager/
[Access Tokens]: https://bitwarden.com/help/access-tokens/
[Personal API Key]: https://bitwarden.com/help/personal-api-key/
[provider aliases]: https://developer.hashicorp.com/terraform/language/providers/configuration#alias-multiple-provider-configurations
