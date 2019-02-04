#!/usr/bin/env bash

set -e
set -o pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 <env-name>"
  echo "  e.g.: $0 beque"
  echo "NOTE: This script does not work with bosh-lites yet."
  exit 1
fi

ENV=$1
CERT_DIR=cf-credentials/cert_dir

source ~/workspace/cli-private/set_int_test_foundation.sh $ENV

cd $GOPATH/src/code.cloudfoundry.org/cli

echo "Making windows..."
make out/cf-cli_winx64.exe

echo "Creating cli binaries..."
mkdir -p cf-cli-binaries/

pushd out > /dev/null
 tar cvzf ../cf-cli-binaries/cf-cli-binaries.tgz cf-cli_winx64.exe
popd > /dev/null

echo "Creating cf credentials..."
mkdir -p cf-credentials/

echo $CF_INT_PASSWORD > cf-credentials/cf-password

echo $CF_INT_OIDC_PASSWORD > cf-credentials/uaa-oidc-password

echo "Creating bosh-lock dir..."
mkdir -p bosh-lock

echo $ENV.cli.fun > bosh-lock/name

echo "Creating cert dirs.."
mkdir -p $CERT_DIR

bosh int ~/workspace/cli-private/ci/infrastructure/$ENV/bbl-state.json --path /lb/cert > "$CERT_DIR/$ENV.lb.cert"

credhub login --skip-tls-validation
credhub get --name /bosh-$ENV/cf/router_ca | bosh interpolate - --path /value/certificate > "$CERT_DIR/$ENV.router.ca"

echo "flying..."
fly -t ci execute -c ci/cli/tasks/integration-windows.yml -i cli=. -i cf-cli-binaries=./cf-cli-binaries -i cli-ci=. -i cf-credentials=./cf-credentials -i bosh-lock=./bosh-lock --tag "cli-windows"

echo "DONE"
