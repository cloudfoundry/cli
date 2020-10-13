#!/bin/bash

CF=$1
export CF_HOME=/tmp/cf_test

for i in {1..100}; do
  $CF api https://api.arsicault.cli.fun --skip-ssl-validation
  $CF api --unset
done
