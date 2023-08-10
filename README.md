# Terraform Provider for Bitwarden

![Tests](https://github.com/maxlaverse/terraform-provider-bitwarden/actions/workflows/tests.yml/badge.svg?branch=main)
[![Coverage Status](https://coveralls.io/repos/github/maxlaverse/terraform-provider-bitwarden/badge.svg?branch=main)](https://coveralls.io/github/maxlaverse/terraform-provider-bitwarden?branch=main)
![Go Version](https://img.shields.io/github/go-mod/go-version/maxlaverse/terraform-provider-bitwarden)
![Releases](https://img.shields.io/github/v/release/maxlaverse/terraform-provider-bitwarden?include_prereleases)


The Terraform Bitwarden provider is a plugin for Terraform that allows to manage different kind of Bitwarden resources.
This project is not associated with the Bitwarden project nor 8bit Solutions LLC.

**[Explore the docs »][Terraform Registry docs]**

---

## Table of Contents
- [Supported Versions](#supported-versions)
- [Usage](#usage)
- [Developing the Provider](#developing-the-provider)
- [License](#license)

## Supported Versions
The plugin has been tested and built with the following components:
- [Terraform] v1.5.2
- [Bitwarden CLI] v2023.7.0
- [Go] 1.20.7 (for development)
- [Docker] 23.0.5 (for development)

The provider likely works with older versions but those haven't been tested.

## Usage

The complete documentation for this provider can be found on the [Terraform Registry docs].

```tf
# Setting up the Provider
variable "bw_password" {
  type        = string
  description = "Bitwarden Master Key"
  sensitive   = true
}

variable "bw_client_id" {
  type        = string
  description = "Bitwarden Client ID"
  sensitive   = true
}

variable "bw_client_secret" {
  type        = string
  description = "Bitwarden Client Secret"
  sensitive   = true
}

terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.5.0"
    }
  }
}

provider "bitwarden" {
  master_password = var.bw_password
  client_id       = var.bw_client_id
  client_secret   = var.bw_client_secret
  email           = "test@laverse.net"
  server          = "https://vault.bitwarden.com"
}


# Managing Folders
resource "bitwarden_folder" "cloud_credentials" {
  name = "My Cloud Credentials"
}


# Managing Logins and Secure Notes
resource "random_password" "vpn_password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "bitwarden_item_login" "vpn_credentials" {
  folder_id = bitwarden_folder.cloud_credentials.id

  name      = "VPN Read Only User/Password Access"
  username  = "vpn-user"
  password  = random_password.vpn_password.result
}

resource "bitwarden_item_secure_note" "vpn_note" {
  folder_id = bitwarden_folder.cloud_credentials.id

  name      = "Notes on the preshared Secret"
  notes     = "It's 1234"
}


# Managing Attachments
resource "bitwarden_attachment" "vpn_config" {
  file = "./vpn_config.txt"
  item_id = bitwarden_item_login.vpn_note.id
}


# Using Login information
data "bitwarden_item_login" "mysql_credentials" {
  id = "ec4e447f-9aed-4203-b834-c8f3848828f7"
}

resource "kubernetes_secret" "database" {
  metadata {
    name = "database"
  }

  data = {
    username = data.bitwarden_item_login.mysql_root_credentials.username
    password = data.bitwarden_item_login.mysql_root_credentials.password
  }
}


# Using Attachments
data "bitwarden_attachment" "ssh_credentials" {
  id = "4d6a41364d6a4dea8ddb1a"
  item_id = "59575167-4d36-5a58-466e-d9021926df8a"
}

resource "kubernetes_secret" "ssh" {
  metadata {
    name = "ssh"
  }

  data = {
    "private.key" = data.bitwarden_attachment.ssh_credentials.content
  }
}
```

See the [examples](./examples/) directory for more examples.

## Security Considerations

The Terraform Bitwarden provider entirely relies on the [Bitwarden CLI] to interact with Vaults.
When you ask Terraform to *plan* or *apply* changes, the provider downloads the encrypted Vault locally as if you would use the Bitwarden CLI directly.
Currently, the Terraform SDK doesn't offer a way to remove the encrypted Vault once changes have been applied.
The issue [hashicorp/terraform-plugin-sdk#63] tracks discussions for adding such a feature.

If you want find out more about this file, you can read [Terraform's documentation on Data Storage].
Please note that this file is stored at `<your-project>/.bitwarden/` by default, in order to not interfer with your local Vaults.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, start a Vaultwarden server:
```sh
$ docker run -ti \
  -e I_REALLY_WANT_VOLATILE_STORAGE=true \
  -e ADMIN_TOKEN=test1234 \
  -e LOGIN_RATELIMIT_SECONDS=1 \
  -e LOGIN_RATELIMIT_MAX_BURST=1000000 \
  -e ADMIN_RATELIMIT_SECONDS=1 \
  -e ADMIN_RATELIMIT_MAX_BURST=1000000 \
  -p 8080:80 vaultwarden/server
```

Then run `make testacc`.

```sh
$ make testacc
```


## License

Distributed under the Mozilla License. See [LICENSE](./LICENSE) for more information.

[Terraform]: https://www.terraform.io/downloads.html
[Go]: https://golang.org/doc/install
[Bitwarden CLI]: https://bitwarden.com/help/article/cli/#download-and-install
[Docker]: https://www.docker.com/products/docker-desktop
[Terraform Registry docs]: https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs
[hashicorp/terraform-plugin-sdk#63]: https://github.com/hashicorp/terraform-plugin-sdk/issues/63
[Terraform's documentation on Data Storage]: https://bitwarden.com/help/data-storage/#on-your-local-machine