# Terraform Provider for Bitwarden

![Tests](https://github.com/maxlaverse/terraform-provider-bitwarden/actions/workflows/tests.yml/badge.svg?branch=main)
[![Coverage Status](https://coveralls.io/repos/github/maxlaverse/terraform-provider-bitwarden/badge.svg?branch=main)](https://coveralls.io/github/maxlaverse/terraform-provider-bitwarden?branch=main)
![Go Version](https://img.shields.io/github/go-mod/go-version/maxlaverse/terraform-provider-bitwarden)
![Releases](https://img.shields.io/github/v/release/maxlaverse/terraform-provider-bitwarden?include_prereleases)
![Downloads](https://img.shields.io/badge/dynamic/json?color=7b42bc&label=Downloads&labelColor=black&logo=terraform&query=data.attributes.total&url=https%3A%2F%2Fregistry.terraform.io%2Fv2%2Fproviders%2F2657%2Fdownloads%2Fsummary&style=flat-square)

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
- [Terraform] v1.6.1 / [OpenTofu] v1.8.0
- [Bitwarden CLI] 2024.7.2
- [Go] 1.22.0 (for development)
- [Docker] 23.0.5 (for development)

The provider likely works with older versions but those haven't been tested.

## Usage

The complete documentation for this provider can be found on the [Terraform Registry docs].

```tf
terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.8.0"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email = "terraform@example.com"
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
  -e DISABLE_ICON_DOWNLOAD=false \
  -e ADMIN_TOKEN=test1234 \
  -e LOGIN_RATELIMIT_SECONDS=1 \
  -e LOGIN_RATELIMIT_MAX_BURST=1000000 \
  -e ADMIN_RATELIMIT_SECONDS=1 \
  -e ADMIN_RATELIMIT_MAX_BURST=1000000 \
  --mount type=tmpfs,destination=/data \
  -p 8080:80 vaultwarden/server
```

Then run `make testacc`.

```sh
$ make testacc
```


## License

Distributed under the Mozilla License. See [LICENSE](./LICENSE) for more information.

[Terraform]: https://www.terraform.io/downloads.html
[OpenTofu]: https://opentofu.org/docs/intro/install/
[Go]: https://golang.org/doc/install
[Bitwarden CLI]: https://bitwarden.com/help/article/cli/#download-and-install
[Docker]: https://www.docker.com/products/docker-desktop
[Terraform Registry docs]: https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs
[hashicorp/terraform-plugin-sdk#63]: https://github.com/hashicorp/terraform-plugin-sdk/issues/63
[Terraform's documentation on Data Storage]: https://bitwarden.com/help/data-storage/#on-your-local-machine
