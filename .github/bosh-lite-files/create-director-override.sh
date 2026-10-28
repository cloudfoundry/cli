#!/bin/sh
# Overrides bbl's generated create-director.sh so we can apply extra ops files
# to the BOSH Lite director. Mirrors the stock bosh-lite-gcp plan-patch override
# and adds director-inotify.yml (raises fs.inotify limits on the director host
# so Noble app Envoy sidecars don't crash with "inotify_fd_ >= 0").
bosh create-env \
 ${BBL_STATE_DIR}/bosh-deployment/bosh.yml \
 --state ${BBL_STATE_DIR}/vars/bosh-state.json \
 --vars-store ${BBL_STATE_DIR}/vars/director-vars-store.yml \
 --vars-file ${BBL_STATE_DIR}/vars/director-vars-file.yml \
 --var-file gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}" \
 -v project_id="${BBL_GCP_PROJECT_ID}" \
 -v zone="${BBL_GCP_ZONE}" \
 -v env_name="${ENV_NAME}" \
 -o ${BBL_STATE_DIR}/bosh-deployment/gcp/cpi.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/jumpbox-user.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/uaa.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/credhub.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/bosh-lite.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/bosh-lite-runc.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/gcp/bosh-lite-vm-type.yml \
 -o ${BBL_STATE_DIR}/bosh-deployment/gcp/director-inotify.yml \
 -o ${BBL_STATE_DIR}/external-ip-gcp.yml \
 -o ${BBL_STATE_DIR}/ip-forwarding.yml
