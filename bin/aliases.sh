#!/bin/bash
#


alias tf='terraform'
alias tfa='terraform apply'
alias tfc='rm -rf backend.tf .terraform .terraform.lock.hcl .terragrunt-cache provider.tf'
alias tfd='terraform destroy'
alias tfi='terraform init -input=false -upgrade'
alias tfl='tflint'
alias tfm='terraform import'
alias tfo='terraform output'
alias tfp='terraform plan'
alias tfv='terraform validate'

alias tg='terragrunt'
alias tgc='rm -rf .terraform.lock.hcl .terragrunt-cache'
alias tga='terragrunt apply'
alias tgd='terragrunt destroy'
alias tgi='terragrunt init -upgrade'
alias tgm='terragrunt import'
alias tgo='terragrunt output'
alias tgp='terragrunt plan'
alias tgv='terragrunt hcl validate --inputs --strict'
