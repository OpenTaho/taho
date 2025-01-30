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

This tool is inspired by the [OpenTofu Style Conventions][1], as well as input
from online communities related to Tofu and Terraform, and opinions of those
contribute to this tool.

This tool initializes, checks and/or restructures Terraform module directories
such that the code is structured as follows.

1. `main.tf` exists
2. `variables.tf` has only variable type blocks
3. `outputs.tf` has only output type blocks
4. `terraform.tf` has only terraform type blocks
5. `providers.tf` has only terraform type blocks
6. When arguments and nested blocks exist within a block body,
   arguments are placed before blocks below them. One blank line to separate the
   arguments from the blocks and one blank line is used to seperate single line
   arguments from multi line arguments. Meta arguments are placed ahead of
   normal arguments. Meta blocks are placed after normal blocks. Items are
   arranged is alphabetic within their respective grouping.

## Problems

This tool is still in early development and as such we have a few problems that
we know about as well as many problems that we don't yet know about. The tool
rewrites all Terraform files in a given directory and works well for most
Terraform syntax with the following exceptions.

1. File header comments may be lost.
2. Comments prior to a block but with a line of whitespace between the comment
and the block may be lost or moved.
3. Comments using `//` or `/*` may be mishandled and/or may crash the program.

When you run this tool it is possible that your terraform files will be altered
in ways that introduce errors and as such please make sure you are under version
control prior to running the tool. After you run the tool make sure to test and
review the project.

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
  https://github.com/OpenTaho/taho/releases/download/v0.0.3/taho-0.0.3-darwin-$(arch)"

chmod +x "$HOME/bin/taho"
```

## Install binary for Linux AMD64

A binary can be installed for Linux.

```zsh
mkdir -p "$HOME/bin"

curl -s -L -o "$HOME/bin/taho" \
  https://github.com/OpenTaho/taho/releases/download/v0.0.3/taho-0.0.3-linux-amd64"

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

[1]: https://opentofu.org/docs/language/syntax/style/