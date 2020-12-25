# Vault Token Helper

This is a very simple implementation of a [Token Helper](https://www.vaultproject.io/docs/commands/token-helper) for HashiCorp Vault. It's primary feature is the ability to store multiple active tokens at once, while also not having any additional dependencies.

## Installation

Select a suitable package archive from the [Releases](https://github.com/freakinhippie/vault-token-helper/releases) page. Either `unzip` the archive file to extract the binary for your OS or install the platform specific package using your system package manager.

## Enable Token Helper

Once installed, enabling it by running:

```sh
vault-token-helper enable
```

This will write a suitable configuration file to `~/.vault` or the path specified by the value of the `VAULT_CONFIG_PATH` environment variable.

### Token Storage

Tokens will be written to `~/.config/vault.d/tokens` in JSON. This directory will be created if it doesn't exist.

## Disable Token Helper

Disable the token helper by running:

```sh
vault-token-helper disable
```

That will delete the configuration file, _but will not remove any active tokens_.
