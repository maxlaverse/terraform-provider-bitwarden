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
- [Client Implementation](#client-implementation)
- [Security Considerations](#security-considerations)
- [Contributing](#contributing)
- [Developing the Provider](#developing-the-provider)
- [License](#license)

## Supported Versions

The plugin has been tested with the following components:

- [Terraform] v1.11.4 / [OpenTofu] v1.12.6
- [Bitwarden CLI] v2025.4.0 (when not using the [embedded client](#client-implementation))
- [BWS CLI] v1.0.0 (when not using the [embedded client](#client-implementation))
- [Vaultwarden] v1.33.2
- official [bitwarden.com] servers v2025.6.2

**Minimum Terraform:** 1.10 or later. **Minimum OpenTofu:** 1.12 or later (Protocol 6). Older CLIs that only speak Protocol 5 are not supported.

The provider is likely to work with older versions of the Bitwarden backends/CLIs, but those haven't necessarily been tested.
If you encounter issues with recent versions of the Bitwarden CLI, consider trying out the [embedded client](#client-implementation).

## Usage

The complete documentation for this provider can be found on the [Terraform Registry docs].
Walkthroughs for [authentication][Auth guide], [client implementation][Client Implementation guide], and [ephemeral resources][Ephemeral Resources guide] are published with each release (also on the [OpenTofu Registry docs]).

### Bitwarden Secret

```tf
terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.23.0"
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
  # See the Client Implementation guide.
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
  note  = "Access key for the example project"
}
```

See the [examples](./examples/) directory for more examples.

### Password Manager

```tf
terraform {
  required_providers {
    bitwarden = {
      source  = "maxlaverse/bitwarden"
      version = ">= 0.23.0"
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
  # See the Client Implementation guide.
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

# or read credentials without persisting them in state
ephemeral "bitwarden_item_login" "example" {
  search = "Example"
}
```

See the [examples](./examples/) directory for more examples.

## Client Implementation

By default the provider shells out to the official [Bitwarden CLI] and [BWS CLI]. Set `client_implementation = "embedded"` to talk to Bitwarden servers directly, without those binaries.

The embedded client is faster and has no extra install, but it is a community project. Trade-offs, including local Vault files and device identifiers, are in the [Client Implementation guide].

## Security Considerations

When using the official CLIs, the provider keeps an encrypted Vault copy under `.bitwarden/` by default, and `bws` may pass secret values on the process command line. Details are in the [Client Implementation guide].

## Contributing

Bug reports, documentation fixes, and pull requests are welcome.

This provider is maintained by a single person. For anything more than a small fix, the problem belongs in an issue and the PR should link to it.

### Issues

- Search existing issues first.
- For bugs, follow the issue template.
- For new resources, schema changes, or anything with more than one reasonable design, open an issue before a PR. That’s where to agree on the problem and the approach.

### Pull requests

Link to an issue that already explains the problem. If there isn’t one, the PR should say:

- **what is painful today**, and in which situation
- **why this approach**, including what you considered instead

Keep PRs focused (one problem per PR), add tests for new behavior, and run `mage docs` if you change schemas or examples.

See [Developing the Provider](#developing-the-provider) for how to build and test locally.

By participating, you agree to follow the [Code of Conduct](.github/CODE_OF_CONDUCT.md).

## Developing the Provider

If you wish to work on the provider, you need the following software:
- [Go] 1.27.3
- [Mage] 1.15.0
- [Docker Compose] 2.35.0

Mage is a make/rake-like build tool using Go. You can list all available targets by running `mage` 

### Building the Provider

To compile the provider, run `mage build`.
You can then instruct Terraform to use your compiled version of the provider by running `mage setup:install`.

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

   b. With the Embedded Client against a local Vaultwarden instance. First start Vaultwarden locally:
   ```sh
   $ mage vaultwarden
   ```
   Then, run the tests:
   ```sh
   $ mage test:integrationPwdVaultwardenWithEmbeddedClient
   ```

   c. With the Bitwarden CLI against a local Vaultwarden instance. First start Vaultwarden locally:
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
[BWS CLI]: https://bitwarden.com/help/secrets-manager-cli/
[Docker]: https://www.docker.com/products/docker-desktop
[Docker Compose]: https://docs.docker.com/compose/install/
[Go]: https://golang.org/doc/install
[Mage]: https://magefile.org/
[OpenTofu]: https://opentofu.org/
[Password Manager]: https://bitwarden.com/products/personal/
[Secrets Manager]: https://bitwarden.com/products/secrets-manager/
[Terraform]: https://www.terraform.io/downloads.html
[Terraform Registry docs]: https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs
[OpenTofu Registry docs]: https://search.opentofu.org/provider/maxlaverse/bitwarden/latest
[Auth guide]: https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs/guides/authentication
[Client Implementation guide]: https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs/guides/client-implementation
[Ephemeral Resources guide]: https://registry.terraform.io/providers/maxlaverse/bitwarden/latest/docs/guides/ephemeral-resources
[Vaultwarden]: https://www.vaultwarden.net/
[bitwarden.com]: https://bitwarden.com
