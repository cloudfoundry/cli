#!/bin/bash
# Force-tears-down a bosh-lite environment directly via the GCP CPI and
# gcloud, bypassing bbl/bosh's graceful deployment-deletion path entirely.
#
# Use this when `bbl down` hangs or fails - e.g. warden bind-mount
# "target is busy" errors during deployment deletion, or SSH-tunnel
# connectivity issues to an unhealthy director/jumpbox. `bosh delete-env`
# only needs the local state file + the CPI (GCP API) to destroy a VM, so it
# works even when the director/jumpbox is unresponsive or already gone.
#
# Usage: force-delete-env.sh <env-name>
# Required env vars: BBL_GCP_SERVICE_ACCOUNT_KEY_PATH, GOOGLE_CLOUD_PROJECT,
#                     BBL_GCP_REGION
# Expects <env-name>/bbl-state to already be downloaded in the cwd.

set -uo pipefail

env_name="$1"
state_dir="${env_name}/bbl-state"
zone="$(jq -r '.gcp.zone' "${state_dir}/bbl-state.json")"

echo "Force-deleting director VM for ${env_name}"
bosh delete-env \
  "${state_dir}/bosh-deployment/bosh.yml" \
  --state "${state_dir}/vars/bosh-state.json" \
  --vars-store "${state_dir}/vars/director-vars-store.yml" \
  --vars-file "${state_dir}/vars/director-vars-file.yml" \
  -o "${state_dir}/bosh-deployment/gcp/cpi.yml" \
  -o "${state_dir}/bosh-deployment/jumpbox-user.yml" \
  -o "${state_dir}/bosh-deployment/uaa.yml" \
  -o "${state_dir}/bosh-deployment/credhub.yml" \
  -o "${state_dir}/bbl-ops-files/gcp/bosh-director-ephemeral-ip-ops.yml" \
  --var-file gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}" \
  -v project_id="${GOOGLE_CLOUD_PROJECT}" \
  -v zone="${zone}" \
  --non-interactive || true

echo "Force-deleting jumpbox VM for ${env_name}"
bosh delete-env \
  "${state_dir}/jumpbox-deployment/jumpbox.yml" \
  --state "${state_dir}/vars/jumpbox-state.json" \
  --vars-store "${state_dir}/vars/jumpbox-vars-store.yml" \
  --vars-file "${state_dir}/vars/jumpbox-vars-file.yml" \
  -o "${state_dir}/jumpbox-deployment/gcp/cpi.yml" \
  --var-file gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}" \
  -v project_id="${GOOGLE_CLOUD_PROJECT}" \
  -v zone="${zone}" \
  --non-interactive || true

tfstate="${state_dir}/vars/terraform.tfstate"
if [ -f "$tfstate" ]; then
  echo "Deleting remaining terraform-managed network resources for ${env_name}"

  jq -r '.resources[] | select(.type=="google_compute_firewall") | .instances[0].attributes.name' "$tfstate" | \
    while read -r name; do gcloud compute firewall-rules delete "$name" --quiet || true; done

  jq -r '.resources[] | select(.type=="google_dns_record_set") | "\(.instances[0].attributes.name)\t\(.instances[0].attributes.managed_zone)\t\(.instances[0].attributes.type)"' "$tfstate" | \
    while IFS=$'\t' read -r name dns_zone rtype; do gcloud dns record-sets delete "$name" --zone "$dns_zone" --type "$rtype" --quiet || true; done

  jq -r '.resources[] | select(.type=="google_compute_route") | .instances[0].attributes.name' "$tfstate" | \
    while read -r name; do gcloud compute routes delete "$name" --quiet || true; done

  jq -r '.resources[] | select(.type=="google_compute_router") | .instances[0].attributes.name' "$tfstate" | \
    while read -r name; do gcloud compute routers delete "$name" --region "${BBL_GCP_REGION}" --quiet || true; done

  jq -r '.resources[] | select(.type=="google_compute_address") | .instances[0].attributes.name' "$tfstate" | \
    while read -r name; do gcloud compute addresses delete "$name" --region "${BBL_GCP_REGION}" --quiet || true; done

  jq -r '.resources[] | select(.type=="google_compute_subnetwork") | .instances[0].attributes.name' "$tfstate" | \
    while read -r name; do gcloud compute networks subnets delete "$name" --region "${BBL_GCP_REGION}" --quiet || true; done

  jq -r '.resources[] | select(.type=="google_compute_network") | .instances[0].attributes.name' "$tfstate" | \
    while read -r name; do gcloud compute networks delete "$name" --quiet || true; done
fi

echo "Force-delete complete for ${env_name}"
