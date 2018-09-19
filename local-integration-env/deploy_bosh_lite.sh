#!/bin/bash

set -e

if [[ "$WORKSPACE" -eq "" ]]; then
  WORKSPACE="$HOME/workspace"
fi

CLI_OPS_DIR=$GOPATH/src/code.cloudfoundry.org/cli/ci/local-integration-env/operations
CLI_BOSHLITE_DIR=$WORKSPACE/cli-lite # created by this script for storing BOSH Lite credentials
BOSH_DEPLOYMENT=$WORKSPACE/bosh-deployment # location where this script clones the cloudfoundry/bosh-deployment repository
CF_DEPLOYMENT=$WORKSPACE/cf-deployment # location where this script clones the cloudfoundry/cf-deployment repository

mkdir -p $WORKSPACE
mkdir -p $CLI_BOSHLITE_DIR

ensure_bosh_cli_installed () {
  if [ -z "$(which bosh)" ]; then
    echo "Please install the bosh cli from https://github.com/cloudfoundry/bosh-cli/releases"
    exit 1
  fi
}

cleanup_vms_and_stemcells () {
  echo "removing old bosh lite"
  # power off any running vms
  vboxmanage list runningvms | \
    awk '{print $1}' | \
    grep -i ^vm- | \
    xargs -n1 -I% vboxmanage controlvm % poweroff

  # delete existing bosh lite vms and stemcells and, if that succeeds, associated bosh state file
  vboxmanage list vms | \
    awk '{print $1}' | \
    grep -e ^sc- -e ^vm- | \
    xargs -r -n 1 -I% vboxmanage unregistervm % --delete && \
    rm -rf $HOME/deployments/vbox
}

setup_git_repositories () { # Takes one argument, which is the directory to clone the repos into
  if [ ! -d $BOSH_DEPLOYMENT ]; then
    echo "cloning cloudfoundry/bosh-deployment to $BOSH_DEPLOYMENT"
    git clone https://github.com/cloudfoundry/bosh-deployment.git
  fi
  pushd $BOSH_DEPLOYMENT
    git pull
  popd

  if [ ! -d $CF_DEPLOYMENT ]; then
    echo "cloning cloudfoundry/cf-deployment to $CF_DEPLOYMENT"
    git clone https://github.com/cloudfoundry/cf-deployment.git
  fi
  pushd $CF_DEPLOYMENT
    git checkout master
    git pull
  popd
}

configure_bosh_environment_access () {
  # Create an environment alias for the bosh lite
  bosh alias-env vbox --ca-cert <(bosh int $CLI_BOSHLITE_DIR/creds.yml --path /director_ssl/ca)

  # Set environment and authentication so that we don't have to pass flags to every bosh command
  export BOSH_ENVIRONMENT=vbox # now that we have an alias we can use the human-friendly name
  export BOSH_CLIENT=admin
  export BOSH_CLIENT_SECRET=`bosh int $CLI_BOSHLITE_DIR/creds.yml --path /admin_password`
}

update_bosh_configs () {
  bosh update-runtime-config $BOSH_DEPLOYMENT/runtime-configs/dns.yml \
    --vars-store=$CLI_BOSHLITE_DIR/runtime-config-vars.yml \
    --name=dns

  bosh \
    update-cloud-config $CF_DEPLOYMENT/iaas-support/bosh-lite/cloud-config.yml \
    -o $CLI_OPS_DIR/cloud-config-internet-required.yml
}

interpolate_and_deploy_cf () {
  # Store the uninterpolated manifest on disk for easier debugging and iterating
  cd $CF_DEPLOYMENT
  bosh interpolate cf-deployment.yml \
    -o operations/use-compiled-releases.yml \
    -o operations/bosh-lite.yml \
    -o operations/test/add-persistent-isolation-segment-diego-cell.yml \
    -o operations/experimental/fast-deploy-with-downtime-and-danger.yml \
    -o operations/use-postgres.yml \
    -o $CLI_OPS_DIR/cli-bosh-lite.yml \
    -o $CLI_OPS_DIR/cli-bosh-lite-uaa-client-credentials.yml \
    -o $CLI_OPS_DIR/disable-rep-kernel-params.yml \
    -o $CLI_OPS_DIR/add-oidc-provider.yml \
    -v system_domain=bosh-lite.com \
    -v cf_admin_password=admin > $CLI_BOSHLITE_DIR/cf-manifest-no-vars.yml

  bosh \
    -n \
    -d cf deploy $CLI_BOSHLITE_DIR/cf-manifest-no-vars.yml \
    --vars-store $CLI_BOSHLITE_DIR/deployment-vars.yml
}

setup_routing_for_bosh_ssh () {
  BOSH_LITE_NETWORK=10.244.0.0
  BOSH_LITE_NETMASK=255.255.0.0
  BOSH_LITE_IP=192.168.50.6

  # Set up virtualbox IP as the gateway to our CF
  if ! route | egrep -q "$BOSH_LITE_NETWORK\\s+$BOSH_LITE_IP\\s+$BOSH_LITE_NETMASK\\s"; then
    sudo route add -net $BOSH_LITE_NETWORK netmask $BOSH_LITE_NETMASK gw $BOSH_LITE_IP
  fi
}

login_to_cf () {
  cf api api.bosh-lite.com --skip-ssl-validation
  cf auth admin admin
  cf enable-feature-flag diego_docker
}

### MAIN BODY OF SCRIPT

ensure_bosh_cli_installed

if [[ $1 == "clean" ]]; then
  cleanup_vms_and_stemcells
fi

setup_git_repositories

if [[ -n "$BOSH_ALL_PROXY" ]]; then
  unset $BOSH_ALL_PROXY # if this is set, the bosh cli will fail to talk to the bosh director because it will try to proxy its traffic through the value of this variable.
fi

export BOSH_ENVIRONMENT=192.168.50.6
export BOSH_NON_INTERACTIVE=true # prevent bosh from issuing y/n prompts

bosh create-env $BOSH_DEPLOYMENT/bosh.yml \
  --state $CLI_BOSHLITE_DIR/state.json \
  -o $BOSH_DEPLOYMENT/virtualbox/cpi.yml \
  -o $BOSH_DEPLOYMENT/virtualbox/outbound-network.yml \
  -o $BOSH_DEPLOYMENT/bosh-lite.yml \
  -o $BOSH_DEPLOYMENT/bosh-lite-runc.yml \
  -o $BOSH_DEPLOYMENT/jumpbox-user.yml \
  -o $CLI_OPS_DIR/bosh-lite-more-power.yml \
  --vars-store $CLI_BOSHLITE_DIR/creds.yml \
  -v director_name="Bosh Lite Director" \
  -v internal_ip=$BOSH_ENVIRONMENT \
  -v internal_gw=192.168.50.1 \
  -v internal_cidr=192.168.50.0/24 \
  -v outbound_network_name=NatNetwork

configure_bosh_environment_access

update_bosh_configs

CFD_STEMCELL_VERSION="$(bosh int $CF_DEPLOYMENT/cf-deployment.yml --path /stemcells/alias=default/version)"
bosh upload-stemcell https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=$CFD_STEMCELL_VERSION

interpolate_and_deploy_cf

setup_routing_for_bosh_ssh

login_to_cf
