# Taho

A tool to make Tofu and Terraform code pretty.

## About the Name

Taho is a dessert made with tofu.

Taho is also the name of this command line tool.

## Overview

The Taho CLI supports Site Reliability Engineers working with Terraform,
OpenTofu, Terragrunt, and AWS. Our CLI provides a higher level wrapper around
Terragrunt, a very powerful formatting command for HCL code, and a managed
Docker based environment with Terraform, OpenTofu, AWS, and related tools.

Our CLI provides the following commands:

|Command  |Description                                |
|---------|-------------------------------------------|
|[Apply]  |Applies Terragrunt units                   |
|[Check]  |Checks Terragrunt units                    |
|[Destroy]|Destroys Terragrunt units                  |
|[Disable]|Disables Terragrunt units                  |
|[Docker] |Starts a Docker managed container          |
|[Enable] |Enables Terragrunt units                   |
|[Format] |Disables Terragrunt units                  |
|[Install]|Installs our go binary and scripting layers|
|[Lint]   |Lint for various tools                     |
|[List]   |Lists Terragrunt units                     |
|[Version]|Shows the version of our tool              |

[Install]: #install-command
[Version]: #version-command
[Format]: #format-command

## Format Command

Our `fmt` command is inspired by [Terraform Best Practices][tf-guide], [OpenTofu
Style Conventions][tu-guide], [Terragrunt Best Practices][tg-guide], as well as
input from online communities related to Tofu and Terraform, and opinions of
those who contribute to this tool. Our format command goes beyond the simple
formatting provided by other tools.

This command formats Terraform module directories such that the code is
structured as follows.

1. `main.tf` exists
2. `variables.tf` has only `variable` type blocks
3. `outputs.tf` has only `output` type blocks
4. `terraform.tf` has only `terraform` type blocks
5. `providers.tf` has only `provider` type blocks
6. `*.hcl` files are formatted such that blocks follow attributes.
7. When arguments and nested blocks exist within a block body,
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

## Install Command

```zsh
git clone https://github.com/OpenTaho/taho.git
cd taho
sudo ./script install
```

The tool can also be invoked with `-v` or `--version` to report it's version.

```zsh
taho --version
```

## Version Command

The version command shows the version of the Taho CLI. This command also has
`-v` as an alias.

[tu-guide]: https://opentofu.org/docs/language/syntax/style
[tf-guide]: https://developer.hashicorp.com/terraform/language/style
[tg-guide]: https://docs.gruntwork.io/2.0/docs/overview/concepts/labels-tags
