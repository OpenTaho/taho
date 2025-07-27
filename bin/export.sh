#!/bin/zsh
#
# Export environment variables

export HISTFILE="$HOME/.history-x"

export TF_PLUGIN_CACHE_DIR="/workspace/.tmp"

export PATH="$HOME/.tfenv/bin:$PATH"
export PATH="$HOME/.tgenv/bin:$PATH"
export PATH="$HOME/bin:$PATH"
export PATH="$HOME/go/bin:$PATH"

taho_precmd() {
  echo -n -e "$(taho get-context)"
}

add-zsh-hook precmd taho_precmd

alias tf='terraform'
alias tfa='terraform apply'
alias tfc='rm -rf .terraform'
alias tfd='terraform destroy'
alias tfi='terraform init -input=false -upgrade'
alias tfl='tflint'
alias tfm='terraform import'
alias tfo='terraform output'
alias tfp='terraform plan -lock=false'
alias tfu='rm -rf .terraform.lock.hcl .terraform'
alias tfui='rm -rf .terraform.lock.hcl .terraform; terraform init -input=false -upgrade'
alias tfv='terraform validate'

alias tg='terragrunt'
alias tga='terragrunt apply'
alias tgc='rm -rf .terragrunt-cache'
alias tgd='terragrunt destroy'
alias tgi='terragrunt init -upgrade'
alias tgm='terragrunt import'
alias tgo='terragrunt output'
alias tgp='terragrunt plan -lock=false'
alias tgu='rm -rf .terraform.lock.hcl .terragrunt-cache'
alias tgup='rm -rf .terraform.lock.hcl .terragrunt-cache; terragrunt init -upgrade; terragrunt plan -lock=false'
alias tgv='terragrunt hcl validate --inputs --strict'