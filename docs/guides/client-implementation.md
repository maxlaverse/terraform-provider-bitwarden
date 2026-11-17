---
page_title: "Client Implementation"
description: |-
  Choose between the official Bitwarden CLIs and the embedded client, and understand local data the provider stores.
---

# Client Implementation

The provider can talk to Bitwarden in two ways. Set `client_implementation` on the provider block, or leave it unset to use the default.

```terraform
provider "bitwarden" {
  # client_implementation = "cli"      # default: official Bitwarden CLIs
  # client_implementation = "embedded" # talk to Bitwarden servers directly
}
```

## Official Bitwarden CLIs (Default)

By default, the provider uses the official Bitwarden command-line tools ([Bitwarden CLI] for Password Manager and [BWS CLI] for Secrets Manager). This approach leverages the battle-tested reliability of Bitwarden's own tooling, backed by their engineering team and security expertise.

The trade-off is that you need to pre-install the appropriate CLI tools in your Terraform environment. Additionally, the Password Manager CLI (requiring Node.js) can create performance bottlenecks when managing many resources due to process spawning overhead.

### Local Vault files

When not using the embedded client, the provider downloads the encrypted Vault locally during _plan_ or _apply_ operations, as the Bitwarden CLI would if you used it directly.
There is currently no plugin hook to remove that encrypted Vault once changes have been applied.

The file is stored at `<your-project>/.bitwarden/` by default, so it does not interfere with other local Vaults. You can change that with `vault_path` (env: `BITWARDENCLI_APPDATA_DIR`). See [Bitwarden's documentation on data storage] for more about this file.

### Secrets Manager CLI arguments

When using the official CLIs instead of the embedded client, `bws secret create` and `bws secret edit` pass secret values on the process command line.

## Embedded Client

The provider also includes an embedded client that communicates directly with Bitwarden servers without external dependencies. Set `client_implementation = "embedded"` to use it. This eliminates the need to install separate CLI tools and provides better performance by avoiding external process spawning, making it particularly beneficial for managing large resource sets.

However, this implementation is developed and maintained by a single person as a community project without company resources. While effort goes into ensuring security and correctness, it lacks the extensive security review, testing infrastructure, and dedicated security team that backs Bitwarden's official tools.

### Device identifier in CI

When using the embedded client in CI or containers, the provider stores a device identifier in `.bitwarden/device_identifier`. If that file is missing, a new one is generated each run and Bitwarden may send "new device logged in" emails. To avoid that, persist `.bitwarden/device_identifier` (e.g. in a secret or volume) and restore it before running Terraform.

## Choosing Your Implementation

The official CLIs leverage Bitwarden's proven tooling. The embedded client is a community project that avoids extra binaries and process spawning. Pick the official CLIs when you want Bitwarden's own tools in the path; pick embedded when you want fewer moving parts in CI or when you manage a large number of resources.

[Bitwarden CLI]: https://bitwarden.com/help/article/cli/#download-and-install
[BWS CLI]: https://bitwarden.com/help/secrets-manager-cli/
[Bitwarden's documentation on data storage]: https://bitwarden.com/help/data-storage/#on-your-local-machine
