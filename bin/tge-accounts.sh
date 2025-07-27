#!/bin/bash
#
# Account selection for Terragrunt Execution scripting

export TABLE_COLS='true'
export TGE_ALIAS_ROOT='true'

case "$TGE_ENVIRONMENT" in
  dev|ivv|stg)
    export AWS_ACCOUNT='483853174698'
    export AWS_ALIAS='identitygsc-nonprd'
    ;;
  poc)
    export AWS_ACCOUNT='597088012807'
    export AWS_ALIAS='idamprivacy-sand'
    ;;
  ppe|prd)
    export AWS_ACCOUNT='160433694605'
    export AWS_ALIAS='identitygsc-prd'
    ;;
  *)
    echo 'Unknown environment' >&2
    exit 1
esac
