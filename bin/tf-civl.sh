#!/bin/bash
#
# Executes a sequence of commands for a Terraform module.
# 
# This is defined as a script rather than as an alias in order to support
# invoking this using sre-dkr. 

if ! terraform -version > /dev/null 2>&1; then
  export PATH="$HOME/.tfenv/bin:$PATH"
fi

rm -rf .terraform && rm -rf .terraform.lock.hcl .terragrunt-cache backend.tf provider.tf
terraform init
terraform validate
tflint
