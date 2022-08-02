mkdir -p go/src/code.cloudfoundry.org
ln -s ${PWD}/cli go/src/code.cloudfoundry.org
ENV=$(cat metadata.json | jq -r '.name')
# export SSL_CERT_DIR=/tmp/cert_dir
# mkdir -p $SSL_CERT_DIR

eval "$(bbl print-env --metadata-file ./metadata.json)"

credhub login
export CF_INT_PASSWORD=$(credhub get -n /bosh-$ENV/cf/cf_admin_password | bosh interpolate --path /value -)
export CF_INT_OIDC_USERNAME="admin-oidc"
export CF_INT_OIDC_PASSWORD=$(credhub get -n /bosh-$ENV/cf/uaa_oidc_admin_password | bosh interpolate --path /value -)

# credhub get --name /bosh-$ENV/cf/router_ca | bosh interpolate - --path /value/certificate > $SSL_CERT_DIR/$ENV.router.ca

echo "Deployed CAPI version:"
bosh -d cf releases | grep capi

# Enable SSL Validation once toolsmiths supports it
# export SKIP_SSL_VALIDATION=false

set -x

export CF_INT_API="https://api.${ENV}.cf-app.com"
export CF_DIAL_TIMEOUT=15
export CF_USERNAME=admin

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

export FLAKE_ATTEMPTS=2
export NODES=16

cd $GOPATH/src/code.cloudfoundry.org/cli
go install github.com/onsi/ginkgo/ginkgo@v1.16.4

make integration-tests-full-ci