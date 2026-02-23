# Taho

A tool for Terraform, Tofu, and AWS.

## About the Name

Taho is a dessert made with tofu. Taho is also a city in Utah known for good ski
resorts.

In addtion, Taho is the name of this subcommand line tool. The author of the
project simply picked the name because it goes well with OpenTofu and it is
short.

## Open Issues

- Documentation is very weak
- GoLang implementation is a mess
- Script implmentatation is also a mess
- The `fmt` command has [problems](#problems) that need to be fixed
- Scripts may not be portable currently development is on `bash@5.2.37` and
  `zsh@5.9`

## Overview

The Taho CLI supports Site Reliability Engineers working with AWS, Docker,
Terraform, Terragrunt, Open Tofu, and Kubernetes, and related tools.

Subcommands to address the followning.

- Formatting for Terraform and Terragrunt files
- Scripting for Terraform, and Terragrunt projects
- Docker Based Shell enviornment

Many features are currently implemented by a large [shell script](./script). If
others contribute to the project the tool will become more useful.

The tool is at the point where it is useful for the work it's author does on
projects.

## Subcommands

Our CLI provides the following subcommands:

|Subcommand Name     |Subcommand     |Description                                                                 |
|--------------------|---------------|----------------------------------------------------------------------------|
|[AWS-RunAs]         |`aws-runas`    |AWS RunAs script output                                                     |
|[Check]             |`check`        |Checks infrastructure units                                                 |
|[Clean]             |`clean`        |Removes Terraform and Terragrunt cache files                                |
|[Copy-Locks]        |`copy-locks`   |Copies the `.terraform.lock.hcl` files from the module to units directories |
|[Destroy]           |`destroy`      |Destroys infrastructure units                                               |
|[Disable]           |`disable`      |Disables Terragrunt units                                                   |
|[Enable]            |`enable`       |Enables Terragrunt units                                                    |
|[Format]            |`fmt`          |Formats code                                                                |
|[GC]                |`gc`           |Collect Garbage                                                             |
|[Init]              |`init`         |Initializes infrastructure units                                            |
|[Install-Kubectl]   |`install-k`    |Installs Kubectl                                                            |
|[Install-Terraform] |`install-tf`   |Installs Terraform                                                          |
|[Install-Terragrunt]|`install-tg`   |Installs Terragrunt                                                         |
|[Install]           |`install`      |Installs Taho CLI                                                           |
|[Lint]              |`lint`         |Lint at the repository level                                                |
|[List]              |`ls`           |Lists infrastructure units                                                  |
|[Save-AWS-AUTH]     |`save-aws-auth`|Saves AWS Authentication in a `.tmp/.aws/$AWS_ALIAS` file                   |
|[Shell]             |`shell`        |Starts a Docker managed shell enviornment                                   |
|[Start]             |`start`        |Starts by adding and/or refining environment and adding aliases             |
|[Tag-Version]       |`tag-v`        |Create a tag based on branch and commit                                     |
|[TF-Lint-Fix]       |`tf-lint-fix`  |Runs `tflint --fix` for all modules                                         |
|[TF-Lint]           |`tf-lint`      |Init, valiation, format checking, and lint                                  |
|[TFG-Lock]          |`tfg-lock`     |Runs `terraform providers lock ...` or `terragrunt run -- provide...`       |
|[Version]           |`version`      |Shows the version of our tool                                               |
|[URL]               |`url`          |Show the https URL for a github origin                                      |

## AWS RunAs Subcommand

The `aws-runas` subcommand outputs a script that can be used in conjunction with
the [aws-runas-cli] CLI. The subcommand is a wrapper around the `aws-runas`
command where we perform `unset AWS_...` on several enviornment variables prior
to invoking `aws-runas` and afterwards it set a few additional enviornment
variables.

The `aws-runas` subcommand takes one poisitonal parameter to identify the
profile that should be used. The default behavior is to create a `1h` session
unless the `TH_AWS_RUNAS_TIME` enviornment variable is set in which case it uses
the value of that enviornment variable.

## Check Subcommand

The `check` subcommand executes `terragrunt plan -lock=false` for all units
selected from the enviornment list. Passing `-filter` with a regular expression
allows you to limit the scope to only units matching the filter. Our scripting
checks the exit code for the Terragrunt subcommand and a `yes` value is placed
in the `DFT` column for any unit that has a non-zero result
_(i.e. has drifted from the planned state)_.

## Clean Subcommand

The `clean` subcommand removes `.terraform` and `.terragrunt-cache` directories
recursivly from the repository root. Adding the `-unlock` option also removes
`.terraform.lock.hcl` files.

## Copy Locks Subcommand

The `copy-locks` subcommand copies the `.terraform.lock.hcl` files from the
module directory into the unit directory.

## Destroy Subcommand

The `destroy` subcommand performs a `terragrunt destroy` for all units selected
from the enviornment list. Passing `-filter` with a regular expression allows
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

The following example will format files in the current directory.

```zsh
taho fmt
```

The following example will format files recursively in all subdirectories.

```zsh
taho fmt -r
```

When you run this tool it is possible that your Terraform and Terragrunt files
will be altered in ways that may introduce errors and as such please make sure
you are under version control prior to running the tool. After you run the tool
make sure to test and review the project.

## GC Subcommand

The `gc` subcommand deletes the `.tmp` directory from the repository root as well
as the `$HOME/.taho/tmp` directory where other commands have likely left
"garbage". This subcommand should be issued when no other subcommands are
running.

## Init Subcommand

The `init` subcommand executes `terraform init` or `terragrunt init` on all
directories associated with an enviornment.

## Install Subcommand

The recommended way to install is to use `sudo` and install with the default
location.

```zsh
git clone https://github.com/OpenTaho/taho
cd taho
sudo ./script install
```

If you do not have `sudo` access you can modify the install subcommand as
follows. If you are taking this approach you will need to also edit your
`.zshrc` file to include the export that modifies the `PATH` variable.

```zsh
mkdir -p "$HOME/bin"
export PATH="$HOME/bin:$PATH"
./script install "$HOME/bin"
```

The tool can also be invoked with `-v` or `--version` to report it's version.

```zsh
taho --version
```

## Install Kubectl Subcommand

The `install-k` subcommand installs a specified Kubectl binary.

## Install Terraform Subcommand

The `install-tf` subcommand installs a specified Terraform binary.

## Install Terragrunt Subcommand

The `install-tg` subcommand installs a specified Terragrunt binary.

## Install TFLint Subcommand

The `install-tflint` subcommand installs a specified Terraform binary.

## Lint Subcommand

The `lint` subcommand performs lint checks from the root of the repository.

## List Subcommand

The `list` subcommand displays a list created based on the `UNITS.md` file in
the root of the repository.

The list subcommand requires a parameter to define the environment that should
be listed. Invoke the list subcommand as shown in the following example.

```zsh
taho list prd
```

## Print Working Directory Subcommand

The `pwd` subcommand prints the current working directory relative to the
repository root.

## Print Unit Directory Subcommand

The `pud` subcommand prints the directory for the unit in the `.tmp` directory.

## Save AWS Auth

The `save-aws-auth` subcommand saves enviornment variables for AWS
authenticiation to a `.tmp/.aws/$AWS_ALIAS` file. The [Check] command will
automatically use these files if it is invoked without an `AWS_ACCESS_KEY_ID`
enviornment variable.

## Shell Subcommand

The `shell` subcommand starts a Docker container for the repository. You can
pass additiona parameter which are executed within the shell.

The `shell` subcommand is the default so issuing the `taho` or `taho shell` will
start a taho shell session.

```zsh
taho shell
```

In the following example, `terraform version` is executed within the Taho shell.

```zsh
taho shell terraform version
```

In the following example, `terraform version` and `terragrunt version` are
executed within the Taho shell. Quotes are required in this case due to how
the semicolon is processed in the zsh shell.

```zsh
taho shell 'terraform version; terragrunt --version'
```

## Start Subcommand

The `start` subcommand outputs a script that defines aliases and shell functions.

Some of the aliases in turn support the [AWS RunAs](#aws-runas-subcommand).

On MacOS, invoke the `start` subcommand as follows:

```zsh
source <(taho start)
```

We recommend adding `alias taho-start='source <(taho start)'` to your `.zshrc`
so it is convenient to issue the subcommand each time you start working in a
repository. The `start` subcommand can potentially incorporate content from your
repositories `.taho.sh` file.

## Tag Version Subcommand

The `tag-v` subcommand creates a tag based on the branch name, the text of the
latest commit, and the history of tags.

The following tokens can be included in either the branch name or in the text of
the latest commit.  If no tokens are found the default behavior is to create a
new tag with an incremented minor component.

|Token       |Behavior                                               |
|------------|-------------------------------------------------------|
|`@major`    |Increment to the major component                       |
|`@minor`    |Increment to the minor component                       |
|`@patch`    |Increment to the patch component                       |
|`@pre-major`|Increment to the major component and pre-release number|
|`@pre-minor`|Increment to the minor component and pre-release number|
|`@pre-patch`|Increment to the patch component and pre-release number|

## TF Lint Fix Subcommand

The `tf-lint-fix` subcommand runs `tflint --fix` for all module directories.

## TF Lint Subcommand

The `tf-lint` subcommand initialize, validate, and runs `tflint` for all
Terraform modules in the repository.

## TFG Lock Subcommand

Runs `terraform providers lock ...` or `terragrunt run -- provide...`.

## URL Subcommand

The `url` subcommand shows the url of the origin. For Github the URL is
converted to https protocol.

## Version Subcommand

The version subcommand shows the version of the Taho CLI. This subcommand also has
`-v` as an alias.

## Configuration

When using this tool with a Terragrunt you also may wan to add `.taho.sh` as a
file in the root of your infrastructure repostory. The `.taho.sh` script will
execute after the Taho subcommand has set the `TAHO_ENVIRONMENT` environment
variable based on user input and before reading your `UNITS.md` file or
accessing AWS.

The configuration approach is intentionally powerful and flexible because it is
able to alter the script enviornment.  Building a good `.taho.sh` script
requires a deep undertanding of the core TahoCLI [script](./script).

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

With a root level `UNITS.md` file for your repository add a table listing the
Terragrunt units that you would like the Taho CLI to consider. This can be a
subset of the full environment list; the tables should take be formatted based
on the following. The Taho CLI will parse the first column. Only the first
column is considered. The `Notes` column in the example is just an example of a
column that is ignored by the tool. The `Notes` column is not required.

```markdown
|`./infrastructure/the-nonprd`|Notes                         |
|-----------------------------|------------------------------|
|`eu-west-2/dev/unit-1`       |                              |
|`eu-west-2/dev/unit-2`       |Unit 2 has yada and dada bits.|
|`us-east-1/dev/unit-3`       |                              |
```

Configuration using a `.taho.json` placed in the directory where Taho is
executed or in `$HOME` is another option. The `.taho.json` file allows you to
configure paths that are ignored.

```json
{
  "ignore": [
    "/Users/UFARNMA/my-project/my-dir",
    "/Users/UFARNMA/my-other-project/my-other-dir",
  ]
}
```

## Problems

This tool is still in early development and as such we have a few problems that
we know about as well as many problems that we don't yet know about. The tool
rewrites all Terraform files in a given directory and works well for most
Terraform syntax with the following exceptions.

1. File header comments are not handled correctly
2. Comments prior to a block but with a line of whitespace between the comment
and the block are not handled correctly
3. Using `//` or `/*` may result in a crash
4. HCL errors will result in a crash

[AWS-RunAs]:          #aws-runas-subcommand
[aws-runas-cli]:      https://github.com/mmmorris1975/aws-runas
[Check]:              #check-subcommand
[Clean]:              #clean-subcommand
[Copy-Locks]:         #copy-locks-subcommand
[Destroy]:            #destroy-subcommand
[Disable]:            #disable-subcommand
[Enable]:             #enable-subcommand
[Format]:             #format-subcommand
[GC]:                 #gc-subcommand
[Init]:               #init-subcommand
[Install-Kubectl]:    #install-kubectl-subcommand
[Install-Terraform]:  #install-terraform-subcommand
[Install-Terragrunt]: #install-terragrunt-subcommand
[Install]:            #install-subcommand
[Lint]:               #lint-subcommand
[List]:               #list-subcommand
[Save-AWS-AUTH]:      #save-aws-auth
[Shell]:              #shell-subcommand
[Start]:              #start-subcommand
[Tag-Version]:        #tag-version-subcommand
[tf-guide]:           https://developer.hashicorp.com/terraform/language/style
[TF-Lint-Fix]:        #tf-lint-fix-subcommand
[TF-Lint]:            #tf-lint-subcommand
[TFG-Lock]:           #tfg-lock-subcommand
[tg-guide]:           https://docs.gruntwork.io/2.0/docs/overview/concepts/labels-tags
[tu-guide]:           https://opentofu.org/docs/language/syntax/style
[URL]:                #url-subcommand
[Version]:            #version-subcommand
