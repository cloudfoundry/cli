ENV=$(cat metadata.json | jq -r '.name')
DATA_DIR=$PWD/cf-data
mkdir -p $DATA_DIR

eval "$(bbl print-env --metadata-file metadata.json)"

credhub login
CF_INT_PASSWORD=$(credhub get -n /bosh-$ENV/cf/cf_admin_password | bosh interpolate --path /value -)
CF_INT_OIDC_PASSWORD=$(credhub get -n /bosh-$ENV/cf/uaa_oidc_admin_password | bosh interpolate --path /value -)

credhub get --name /bosh-$ENV/cf/router_ca | bosh interpolate - --path /value/certificate > $DATA_DIR/$ENV.router.ca

echo "Deployed CAPI version:"
bosh -d cf releases | grep capi

# set -x

# output password into a temp file for consumption by Windows
echo $CF_INT_PASSWORD > $DATA_DIR/cf-password
echo $CF_INT_OIDC_PASSWORD > $DATA_DIR/uaa-oidc-password

echo "::set-output name=data::$DATA_DIR"