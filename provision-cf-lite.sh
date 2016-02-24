#!/bin/bash

set -e -x

CLI_CI=$PWD/ci

git config --global user.email "ci@localhost"
git config --global user.name "CI Bot"

function commit_vagrant_artifacts() {
  cd ${CLI_CI}

  if [ -n "$(git status --porcelain)" ]; then
    git add -A
    git commit -m "provisioned"
  fi
}

mkdir scratch

pushd scratch
  if [ -d "$VAGRANT_DOTFILE_PATH" ]; then
    vagrant destroy -f
  fi

  vagrant init cloudfoundry/bosh-lite

  sed -i -e "s/do |config|/do |config|\n  config.vm.provider 'aws' do |aws|\n    aws.private_ip_address = '${LITE_PRIVATE_IP_ADDRESS}'\n  end/" \
    Vagrantfile

  vagrant up --provider aws
popd

trap commit_vagrant_artifacts EXIT

sed -i -e "s/bosh-lite-ip-${LITE_NAME}: \(.*\)/bosh-lite-ip-${LITE_NAME}: ${DIRECTOR_IP}/" \
  ${CLI_CI}/concourse/credentials.yml

pushd cf-release
  sed -i -e "s/^properties:/properties:\n  domain: ${LITE_HOSTNAME}/" \
    bosh-lite/stubs/property_overrides.yml
popd

cd bosh-lite

./bin/provision_cf
