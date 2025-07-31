# Taho

A tool to make Tofu and Terraform code pretty.

## About the Name

Taho is a dessert made with tofu. Taho is also a city in Utah known for good ski
resorts.

In addtion, Taho is the name of this subcommand line tool. The founder of the
project simply picked the name because it goes well with OpenTofu and it is
short.

## Overview

The Taho CLI supports Site Reliability Engineers working with Terraform,
OpenTofu, Terragrunt, and AWS. Our CLI provides a higher level wrapper around
Terragrunt, a very powerful formatting subcommand for HCL code, and a managed
Docker based environment with Terraform, OpenTofu, AWS, and related tools.

Currently this tool is in early development. The tool is at the point where it
is somewhat useful for Terraform and Terragrunt projects. Features related to
OpenTofu do not yet work.

## Configuration

When using this tool with a Terragrunt you will need to add `.taho.sh` as a file
in the root of your infrastructure repostory. The `.taho.sh` script will execute
after the Taho subcommand has set the `TAHO_ENVIRONMENT` environment variable based
on user input and before reading your `README.md` file or accessing AWS.

Building a powerful `.taho.sh` script requires a deep undertanding of the core
TahoCLI [script](./script).

```bash
#!/bin/bash
#
# Example Taho repostory, adjust this as needed.
case "$TAHO_ENVIRONMENT" in
  dev|stg)
    export AWS_ACCOUNT='100000000000'
    export AWS_ALIAS='the-nonprd'
    ;;
  prd)
    export AWS_ACCOUNT='100000000001'
    export AWS_ALIAS='the-prod'
    ;;
esac
```

Within the root level `README.md` file for your repository add tables that list
the Terragrunt units that you would like the Taho CLI to consider. This can be a
subset of the full environment list; the tables should take be formatted based
on the following. The Taho CLI will parse the first column. Only the first
column is considered. The `Notes` column in the example is just an example of a
column that is ignored by the tool.

```markdown
|`./infrastructure/the-nonprd`|Notes                         |
|-----------------------------|------------------------------|
|`eu-west-2/dev/unit-1`       |                              |
|`eu-west-2/dev/unit-2`       |Unit 2 has yada and dada bits.|
|`us-east-1/dev/unit-3`       |                              |
```

## Subcommands

Our CLI provides the following subcommands:

|Command  |Description                                  |
|---------|---------------------------------------------|
|[Apply]  |Applies infrastructure units                 |
|[Check]  |Checks infrastructure units                  |
|[Clean]  |Removes Terraform and Terragrunt cache files |
|[Destroy]|Destroys infrastructure units                |
|[Disable]|Disables Terragrunt units                    |
|[Enable] |Enables Terragrunt units                     |
|[Format] |Formats code                                 |
|[Init]   |Initializes infrastructure units             |
|[Install]|Installs our go binary and scripting layers  |
|[Lint]   |Lint at the repository level                 |
|[List]   |Lists infrastructure units                   |
|[Shell]  |Starts a Docker managed container with bash  |
|[Version]|Shows the version of our tool                |

[Apply]: #apply-subcommand
[Check]: #check-subcommand
[Clean]: #clean-subcommand
[Destroy]: #destroy-subcommand
[Disable]: #disable-subcommand
[Enable]: #enable-subcommand
[Format]: #format-subcommand
[Init]: #init-subcommand
[Install]: #install-subcommand
[Lint]: #lint-subcommand
[Shell]: #shell-subcommand
[Version]: #version-subcommand

## Apply Subcommand

The `apply` subcommand performs a `terragrunt apply` for all units selected from
the enviornment list. Passing `-fitler` with a regular expression allows you to
limit the scope to only units matching the filter. The default filter is `.*`.
Passing `-auto-approve` is required if you wish for the apply to proceed with
automatic approval.

## Check Subcommand

The `check` subcommand executes `terragrunt plan -lock=false` for all units
selected from the enviornment list. Passing `-fitler` with a regular expression
allows you to limit the scope to only units matching the filter. The exit code
for the Terragrunt subcommand and a `yes` value is placed in the `DFT` column for
any unit that has a non-zero result _(i.e. has drifted from the planned state)_.

## Clean Subcommand

The `clean` subcommand removes `.terraform` and `.terragrunt-cache` directories
recursivly from the repository root. Adding the `-unlock` option also removes
`.terraform.lock.hcl` files.

## Destroy Subcommand

The `destroy` subcommand performs a `terragrunt destroy` for all units selected
from the enviornment list. Passing `-fitler` with a regular expression allows
you to limit the scope to only units matching the filter. The default filter is
`.*`.  Passing `-auto-approve` is required if you wish for the apply to proceed
with automatic approval.

## Disable Subcommand

The `disable` subcommand alters Terragrunt units setting
`unit = { enable = false }`.

## Enable Subcommand

The `disable` subcommand alters Terragrunt units setting
`unit = { enable = true }`.

## Format Subcommand

The `fmt` subcommand is inspired by [Terraform Best Practices][tf-guide],
[OpenTofu Style Conventions][tu-guide], [Terragrunt Best Practices][tg-guide],
as well as input from online communities related to Tofu and Terraform, and
opinions of those who contribute to this tool. Our format subcommand goes beyond
the simple formatting provided by other tools.

This subcommand formats Terraform module directories such that the code is
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

## Init Subcommand

The `init` subcommand executes `terraform init` or `terragrunt init` on all
directories associated with an enviornment.

## Install Subcommand

```zsh
git clone https://github.com/OpenTaho/taho.git
cd taho
sudo ./script install
```

The tool can also be invoked with `-v` or `--version` to report it's version.

```zsh
taho --version
```

## Lint Subcommand

The `lint` subcommand performs lint checks from the root of the repository.

## Shell Subcommand

The `shell` subcommand starts a Docker container for the repository.

## Version Subcommand

The version subcommand shows the version of the Taho CLI. This subcommand also has
`-v` as an alias.

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

[tu-guide]: https://opentofu.org/docs/language/syntax/style
[tf-guide]: https://developer.hashicorp.com/terraform/language/style
[tg-guide]: https://docs.gruntwork.io/2.0/docs/overview/concepts/labels-tags
