#!/bin/bash

set -eu

# ENV
: "${MIN_UNCLAIMED_COUNT:?}"
: "${POOL_NAME:="bosh-lites"}"
: "${BUILDING_POOL_NAME:="building-bosh-lites"}"
: "${GIT_USERNAME:="CLI CI"}"
: "${GIT_EMAIL:="cf-cli-eng+ci@pivotal.io"}"
: "${TRIGGER_FILE_NAME:=".trigger-bosh-lites-create"}"

# INPUTS
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
workspace_dir="$( cd "${script_dir}/../../../" && pwd )"
pool_dir="${workspace_dir}/env-pool"

# OUTPUTS
output_dir="${workspace_dir}/updated-env-pool"

git clone "${pool_dir}" "${output_dir}"

pushd "${output_dir}" > /dev/null
  echo "Searching for bosh-lites..."

  ready_count="$(find "${POOL_NAME}/unclaimed" -not -path '*/\.*' -type f | wc -l)"
  echo "Ready bosh-lites: ${ready_count}"
  building_count="$(find "${BUILDING_POOL_NAME}/claimed" -not -path '*/\.*' -type f | wc -l)"
  echo "Building bosh-lites: ${building_count}"

  env_count=$((ready_count + building_count))
  echo "Total count: ${env_count}"

  if [ "${env_count}" -lt "${MIN_UNCLAIMED_COUNT}" ]; then
    echo "Fewer than ${MIN_UNCLAIMED_COUNT} bosh-lites, going to trigger creation"
    # The create-bosh-lite job watches this file for changes
    date +%s > $TRIGGER_FILE_NAME

    git config user.name "${GIT_USERNAME}"
    git config user.email "${GIT_EMAIL}"
    git add $TRIGGER_FILE_NAME
    git commit -m "Not enough unclaimed envs in ${POOL_NAME} or ${BUILDING_POOL_NAME} pools, updating trigger."
  fi
popd > /dev/null

echo "DONE"
