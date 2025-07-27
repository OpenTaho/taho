#!/bin/bash
#

alias tf='terraform'
alias tfc='rm -rf .terraform && rm -rf .terraform.lock.hcl .terragrunt-cache backend.tf provider.tf'
alias tfi='terraform init -input=false -upgrade'
alias tfl='tflint'
alias tfm='terraform import'
alias tfv='terraform validate'

alias tg='terragrunt'
alias tga='terragrunt apply'
alias tgd='terragrunt destroy'
alias tgi='terragrunt init -upgrade'
alias tgm='terragrunt import'
alias tgo='terragrunt output'
alias tgp='terragrunt plan'
