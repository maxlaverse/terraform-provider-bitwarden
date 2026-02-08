# Terraform Provider for Bitwarden

![Tests](https://github.com/maxlaverse/terraform-provider-bitwarden/actions/workflows/tests.yml/badge.svg?branch=main)
[![Coverage Status](https://coveralls.io/repos/github/maxlaverse/terraform-provider-bitwarden/badge.svg?branch=main)](https://coveralls.io/github/maxlaverse/terraform-provider-bitwarden?branch=main)
![Go Version](https://img.shields.io/github/go-mod/go-version/maxlaverse/terraform-provider-bitwarden)
![Releases](https://img.shields.io/github/v/release/maxlaverse/terraform-provider-bitwarden?include_prereleases)
![Downloads](https://img.shields.io/badge/dynamic/json?color=7b42bc&label=Downloads&labelColor=black&logo=terraform&query=data.attributes.total&url=https%3A%2F%2Fregistry.terraform.io%2Fv2%2Fproviders%2F2657%2Fdownloads%2Fsummary&style=flat-square)

A provider for Terraform/OpenTofu to manage Bitwarden [Password Manager] and [Secrets Manager] resource.
This project is not associated with the Bitwarden project nor Bitwarden, Inc.

**[Explore the docs on Terraform»][Terraform Registry docs]** &nbsp;&nbsp; or &nbsp;&nbsp; **[Explore the docs on OpenTofu»][OpenTofu Registry docs]**

---

## Table of Contents

- [Supported Versions](#supported-versions)
- [Usage](#usage)
- [Embedded Client](#embedded-client)
- [Security Considerations](#secutiry-considerations)
- [Developing the Provider](#developing-the-provider)
- [License](#license)

## Supported Versions

The plugin has been tested with the following components:

- [Terraform] v1.11.4 / [OpenTofu] v1.10.0
- [Bitwarden CLI] v2025.4.0 (when not using the [Embedded Client](#embedded-client))
- [BWS CLI] v1.0.0 (when not using the [Embedded Client](#embedded-client))
- [Vaultwarden] v1.33.2
- official [bitwarden.com] servers v2025.6.2

The provider is likely to work with older versions, but those haven't necessarily been tested.
If you encounter issues with recent versions of the Bitwarden CLI, consider trying out the [Embedded Client](#embedded-client).

## Usage

The complete documentation for this provider can be found on the [Terraform Registry docs].

### Bitwarden Secret

```tf
terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.17.3"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  access_token = "0.client_id.client_secret:dGVzdC1lbmNyeXB0aW9uLWtleQ=="

  # By default, the provider uses Bitwarden CLIs to interact with the remote
  # Vaults. You can also use the embedded client instead, which removes
  # the need for locally installed binaries.
  #
  # Learn more about the implications by reading the "Client Implementation"
  # section below.
  #
  # client_implementation = "embedded"
}

# Source a project
data "bitwarden_project" "example" {
  id = "37a66d6a-96c1-4f04-9a3c-b1fc0135669e"
}

# Create a Secret
resource "bitwarden_secret" "example" {
  project_id = data.bitwarden_project.example.id

  key   = "ACCESS_KEY"
  value = "THIS-VALUE"
}
```

See the [examples](./examples/) directory for more examples.

### Password Manager

```tf
terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.17.3"
    }
  }
}

# Configure the Bitwarden Provider
provider "bitwarden" {
  email = "terraform@example.com"

  # Specify a server URL when using a self-hosted version of Bitwarden
  # or similar (e.g., Vaultwarden), or the European instances:
  #
  # server = "https://vault.bitwarden.eu"

  # By default, the provider uses Bitwarden CLIs to interact with the remote
  # Vaults. You can also use the embedded client instead, which removes
  # the need for locally installed binaries.
  #
  # Learn more about the implications by reading the "Client Implementation"
  # section below.
  #
  # client_implementation = "embedded"
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

## Client Implementation

The Bitwarden provider offers two client implementations to interact with your Vault:

### Official Bitwarden CLIs (Default)
By default, the provider uses the official Bitwarden command-line tools ([Bitwarden CLI] for Password Manager and [BWS CLI] for Secrets Manager). This approach leverages the battle-tested reliability of Bitwarden's own tooling, backed by their engineering team and security expertise.

The trade-off is that you need to pre-install the appropriate CLI tools in your Terraform environment. Additionally, the Password Manager CLI (requiring Node.js) can create performance bottlenecks when managing many resources due to process spawning overhead.

### Embedded Client
The provider also includes an embedded client that communicates directly with Bitwarden servers without external dependencies. This eliminates the need to install separate CLI tools and provides better performance by avoiding external process spawning, making it particularly beneficial for managing large resource sets.

However, this implementation is developed and maintained by a single person as a community project without company resources. While effort goes into ensuring security and correctness, it lacks the extensive security review, testing infrastructure, and dedicated security team that backs Bitwarden's official tools.

### Choosing Your Implementation

The choice depends on your needs: the official CLIs leverage Bitwarden's proven tooling, while the embedded client is a community project offering performance benefits and zero external dependencies. The embedded client aims for security and correctness, and code reviews are always welcome to help improve it.


## Security Considerations

When not using the [Embedded Client](#embedded-client), the provider downloads the encrypted Vault locally during _plan_ or _apply_ operations as would the Bitwarden CLI if you used it directly.
Currently, the Terraform SDK doesn't offer a way to remove the encrypted Vault once changes have been applied.
The issue [hashicorp/terraform-plugin-sdk#63] tracks discussions for adding such a feature.

If you want find out more about this file, you can read [Terraform's documentation on Data Storage].
Please note that this file is stored at `<your-project>/.bitwarden/` by default, in order to not interfere with your local Vaults.

## Developing the Provider

If you wish to work on the provider, you need the following software:
- [Go] 1.24.1
- [Mage] 1.15.0
- [Docker Compose] 2.35.0

Mage is a make/rake-like build tool using Go. You can list all available targets by running `mage` 

### Building the Provider

To compile the provider, run `mage build`.
You can then instruct Terraform use your compiled version of the provider by running `mage setup:install`.

### Running Tests

The provider includes several types of tests:

1. **Offline Tests**: Run tests that don't require a Bitwarden backend (e.g. schema validation, unit tests):
   ```sh
   $ mage test:offline
   ```

2. **Integration Tests**: There are three types of integration tests:

   a. With the Embedded Client against the official Bitwarden instance. This requires Bitwarden credentials and object identifiers in `.env.official` file (see [.env.official-example](./.env.official-example)):

   ```sh
   $ mage test:integrationPwdOfficialWithEmbeddedClient
   ```

   b. With the Embedded Client against a local Vaultwarden instance. First ensure start Vaultwarden locally:
   ```sh
   $ mage vaultwarden
   ```
   Then, run the tests:
   ```sh
   $ mage test:integrationPwdVaultwardenWithEmbeddedClient
   ```

   c. With the Bitwarden CLI against a local Vaultwarden instance. First ensure start Vaultwarden locally:
   ```sh
   $ mage vaultwarden
   ```
   Then, run the tests:
   ```sh
   $ mage test:integrationPwdVaultwardenWithCLI
   ```

3. **Run All Tests**: To run the complete test suite:
   ```sh
   $ mage test:all
   ```

To clean up test artifacts and clear the test cache:

```sh
$ mage clean
```

### Documentation

To generate or update documentation, run:

```sh
$ mage docs
```

## Disclaimer

While we strive to ensure the reliability and security of this application, we cannot be held liable for any data loss, unauthorized access, or breaches that may result from the way data is stored or processed by third-party providers. This includes, but is not limited to, unencrypted storage of passwords, attachments, and other user information.

Users are strongly encouraged to perform regular backups of their files and database to mitigate potential data loss. In the event of a loss or security concern, please contact us as soon as possible so we can assist where feasible.

## License

Distributed under the Mozilla License. See [LICENSE](./LICENSE) for more information.

[Bitwarden CLI]: https://bitwarden.com/help/article/cli/#download-and-install
[BWS CLI]: https://github.com/bitwarden/sdk-sm/releases
[Docker]: https://www.docker.com/products/docker-desktop
[Docker Compose]: https://docs.docker.com/compose/install/
[Go]: https://golang.org/doc/install
[hashicorp/terraform-plugin-sdk#63]: https://github.com/hashicorp/terraform-plugin-sdk/issues/63
[Mage]: https://magefile.org/
[OpenTofu]: https://opentofu.org/
[Password Manager]: https://bitwarden.com/products/personal/
[Secrets Manager]: https://bitwarden.com/products/secrets-manager/
[Terraform]: https://www.terraform.io/downloads.html
[Terraform Registry docs]: https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs
[OpenTofu Registry docs]: https://search.opentofu.org/provider/maxlaverse/bitwarden/latest
[Terraform's documentation on Data Storage]: https://bitwarden.com/help/data-storage/#on-your-local-machine
[Vaultwarden]: https://www.vaultwarden.net/
[bitwarden.com]: https://bitwarden.com
