#!/bin/zsh
#
# Export environment variables

export HISTFILE="$HOME/.history-x"

export TF_PLUGIN_CACHE_DIR="/workspace/.tmp"

export PATH="$HOME/bin:$PATH"
export PATH="$HOME/go/bin:$PATH"

taho_precmd() {
  echo -n -e "$(taho get-context)"
}

add-zsh-hook precmd taho_precmd

alias gap='git add -p'
alias gca='git commit --amend'
alias gcm='taho-start; taho-git-commit'
alias gdp='taho-git-dpush'
alias gpf='git push -f'
alias gph='git push -u origin HEAD'
alias gpo='git push origin'
alias gpom='git pull; git pull origin main'
alias grm='git rebase -i -- origin/main'
alias gru='git remote update'
alias gst='taho-start; git status'
alias gtd='git tag -d $(git tag); git fetch --tags'
alias gtop='cd "$(git rev-parse --show-toplevel)"'
alias gum='taho-start; export TAHO_COMMIT_TYPE=update; taho-git-commit'
alias gwip='git commit -m wip'

alias taho-start='source <(taho start)'

alias tf='terraform'
alias tfa='terraform apply'
alias tfc='rm -rf .terraform'
alias tfd='terraform destroy'
alias tfi='terraform version; rm -rf .terraform; terraform init -upgrade; terraform validate; tflint; taho tfg-lock; taho fmt'
alias tfl='tflint'
alias tfm='terraform import'
alias tfo='terraform output'
alias tfp='terraform plan -lock=false'
alias tfsm='terraform state mv'
alias tfv='terraform validate'

alias tg='terragrunt'
alias tga='terragrunt apply'
alias tgc='rm -rf .terragrunt-cache'
alias tgd='terragrunt destroy'
alias tgi='terragrunt --version; terraform version; rm -rf .terragrunt-cache; terragrunt init -upgrade; taho tfg-lock'
alias tgm='terragrunt import'
alias tgo='terragrunt output'
alias tgp='terragrunt plan -lock=false'
alias tgv='terragrunt hcl validate --inputs --strict'
