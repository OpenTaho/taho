# Taho

A tool to make Tofu and Terraform code pretty.

## About the name

Taho is a dessert made with tofu.

Taho is also the name of the command line tool we provide to format Tofu and
Terraform modules.

## Current state of the tool

This tool is in development and has not yet reached the point where the first
minimal viable version exists. The tool currently reports it's version as
`0.0.0`.

## Tool Description

This tool is inspired by the [Hashicorp Terraform Style Guide][0], [OpenTofu
Style Conventions][1], input from online communities related to Tofu and
Terraform, the operation of the tools, and opinions of those contribute to this
tool.

This tool initializes, checks and/or restructures Terraform module directories
such that the code is structured as follows.

1. `main.tf` exists
2. `variables.tf` exists with only variable type blocks
3. `outputs.tf` exits with only output type blocks
4. `terraform.tf` exits with only terraform type blocks
5. Blocks are in sorted order

## Install with Go

The tool can installed using Go.

```zsh
git clone https://github.com/OpenTaho/taho.git
cd taho
go install
```

## Install binary for MacOS

A binary can be installed for MacOS.

```zsh
mkdir -p "$HOME/bin"

curl -s -L -o "$HOME/bin/taho" \
  https://github.com/OpenTaho/taho/releases/download/v0.0.1/taho-0.0.1-darwin-$(arch)"

chmod +x "$HOME/bin/taho"
```

## Install binary for Linux AMD64

A binary can be installed for Linux.

```zsh
mkdir -p "$HOME/bin"

curl -s -L -o "$HOME/bin/taho" \
  https://github.com/OpenTaho/taho/releases/download/v0.0.1/taho-0.0.1-linux-amd64"

chmod +x "$HOME/bin/taho"
```

## Install binary for Linux ARM64

The procedure to install for Linux ARM 64 is essentially the same as the
procedure for AMD64 but obvoiusly you must use the suffix `-arm64` for the URL.

## Usage

The default behavior of the tool is to fix the content of the current directory.
If changes are made the tool will output messages using the standard error
stream. The exit code will be  `1` if changes are made or `0` if no changes are
made.

```zsh
taho
```

The tool can also be invoked with `-v` or `--version` to report it's version.

```zsh
taho --version
```

[0]: https://developer.hashicorp.com/terraform/language/style
[1]: https://opentofu.org/docs/language/syntax/style/
