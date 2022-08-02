export GOFLAGS="-mod=mod"
echo "GOFLAGS=$GOFLAGS" >> $GITHUB_ENV

eval "$(bbl print-env --metadata-file metadata.json)"
bosh -d cf manifest > bosh_manifest.yml
bosh interpolate bosh_manifest.yml \
-o cf-deployment/operations/test/add-persistent-isolation-segment-diego-cell.yml \
-o cf-deployment/operations/use-internal-lookup-for-route-services.yml \
-o cf-deployment/operations/use-compiled-releases.yml \
-o cli-ci/ci/infrastructure/operations/add-dummy-windows-stack.yml \
-o cli-ci/ci/infrastructure/operations/enable-mysql-audit-logging.yml \
-o cli-ci/ci/infrastructure/operations/default-app-memory.yml \
-o cli-ci/ci/infrastructure/operations/add-oidc-provider.yml \
-o cli-ci/ci/infrastructure/operations/add-uaa-client-credentials.yml \
-o cli-ci/ci/infrastructure/operations/cli-isolation-cell-overrides.yml \
-o cli-ci/ci/infrastructure/operations/diego-cell-instances.yml \
-o cli-ci/ci/infrastructure/operations/doppler-instances.yml \
-o cli-ci/ci/infrastructure/operations/enable-v3-deployments-endpoint.yml \
-o cli-ci/ci/infrastructure/operations/increase-route-registration-interval.yml \
-o cli-ci/ci/infrastructure/operations/reduce-async-service-polling.yml \
-o cli-ci/ci/infrastructure/operations/adjust-user-retry-attempts.yml \
-o cli-ci/ci/infrastructure/operations/use-latest-capi.yml \
-v client-secret="${CLIENT_SECRET}" \
> ./director.yml

bosh -d cf deploy director.yml -n

echo "Deployed CAPI version:"
bosh -d cf releases | grep capi   