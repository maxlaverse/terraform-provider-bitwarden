# Terraform Provider for Bitwarden

![Tests](https://github.com/maxlaverse/terraform-provider-bitwarden/actions/workflows/tests.yml/badge.svg?branch=main)
[![Coverage Status](https://coveralls.io/repos/github/maxlaverse/terraform-provider-bitwarden/badge.svg?branch=main)](https://coveralls.io/github/maxlaverse/terraform-provider-bitwarden?branch=main)
![Go Version](https://img.shields.io/github/go-mod/go-version/maxlaverse/terraform-provider-bitwarden)
![Releases](https://img.shields.io/github/v/release/maxlaverse/terraform-provider-bitwarden?include_prereleases)
![Downloads](https://img.shields.io/badge/dynamic/json?color=7b42bc&label=Downloads&labelColor=black&logo=terraform&query=data.attributes.total&url=https%3A%2F%2Fregistry.terraform.io%2Fv2%2Fproviders%2F2657%2Fdownloads%2Fsummary&style=flat-square)

A provider for Terraform/OpenTofu to manage Bitwarden [Password Manager] and [Secrets Manager] resource.
This project is not associated with the Bitwarden project nor Bitwarden, Inc.

**[Explore the docs »][Terraform Registry docs]**

---

## Table of Contents

- [Supported Versions](#supported-versions)
- [Usage](#usage)
- [Embedded Client](#embedded-client)
- [Security Considerations](#secutiry-considerations)
- [Developing the Provider](#developing-the-provider)
- [License](#license)

## Supported Versions

The plugin has been tested and built with the following components:

- [Terraform] v1.9.8 / [OpenTofu] v1.9.0
- [Bitwarden CLI] v2023.2.0 (when not using the [Embedded Client](#embedded-client))
- [Go] 1.23.4 (for development)
- [Docker] 24.0.6 (for development)

The provider is likely to work with older versions, but those haven't been tested.
If you encounter issues with recent versions of the Bitwarden CLI, consider trying out the [Embedded Client](#embedded-client).

## Usage

The complete documentation for this provider can be found on the [Terraform Registry docs].

```tf
terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.13.5"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email = "terraform@example.com"

  # Specify a server url when using a self-hosted version of Bitwarden
  # or similar (e.g. Vaultwarden).
  #
  # server = "https://vault.myserver.org"

  # If you have the opportunity, you can try out the embedded client which
  # removes the need for a locally installed Bitwarden CLI. Please note that
  # this feature is still considered experimental and not recommended for
  # production use yet.
  #
  # experimental {
  #   embedded_client = true
  # }
}

# Create a Bitwarden Login item
resource "bitwarden_item_login" "example" {
  name     = "Example"
  username = "service-account"
  password = "<sensitive>"
}

# or use an existing Bitwarden resource
data "bitwarden_item_login" "example" {
  search = "Example"
}
```

See the [examples](./examples/) directory for more examples.

## Embedded Client

Since version 0.9.0, the provider contains an embedded client that can directly interact with Bitwarden's API, removing the need for a locally installed Bitwarden CLI.
The embedded client makes the provider faster, easier to use, but it still requires more testing.
For now, a feature flag needs to be set in order to use it (`experimental.embedded_client`), with the goal of having it the default in v1.0.0.

## Security Considerations

When not using the [Embedded Client](#embedded-client), the provider downloads the encrypted Vault locally during _plan_ or _apply_ operations as would the Bitwarden CLI if you used it directly.
Currently, the Terraform SDK doesn't offer a way to remove the encrypted Vault once changes have been applied.
The issue [hashicorp/terraform-plugin-sdk#63] tracks discussions for adding such a feature.

If you want find out more about this file, you can read [Terraform's documentation on Data Storage].
Please note that this file is stored at `<your-project>/.bitwarden/` by default, in order to not interfere with your local Vaults.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, start a Vaultwarden server:

```sh
$ make server
```

Then run `make testacc`.

```sh
$ make testacc
```

## License

Distributed under the Mozilla License. See [LICENSE](./LICENSE) for more information.

[Bitwarden CLI]: https://bitwarden.com/help/article/cli/#download-and-install
[Docker]: https://www.docker.com/products/docker-desktop
[Go]: https://golang.org/doc/install
[hashicorp/terraform-plugin-sdk#63]: https://github.com/hashicorp/terraform-plugin-sdk/issues/63
[OpenTofu]: https://opentofu.org/
[Password Manager]: https://bitwarden.com/products/personal/
[Secrets Manager]: https://bitwarden.com/products/secrets-manager/
[Terraform]: https://www.terraform.io/downloads.html
[Terraform Registry docs]: https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs
[Terraform's documentation on Data Storage]: https://bitwarden.com/help/data-storage/#on-your-local-machine
