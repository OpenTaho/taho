#!/bin/zsh
#
# Export environment variables

export HISTFILE="$HOME/.history-x"

export TF_PLUGIN_CACHE_DIR="/workspace/.tmp"

export PATH="$HOME/bin:$PATH"
export PATH="/usr/local/go/bin:$PATH"
export PATH="$HOME/go/bin:$PATH"

taho_precmd() {
  echo -n -e "$(taho get-context)"
}

add-zsh-hook precmd taho_precmd

source <(taho start)
