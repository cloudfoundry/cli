# CLI Workflows

## Units
Units runs four jobs, `shared-values`, `lint`, `units` and `units-windows` on all pushes.

- `shared-values`
  - This determines the go version from `go.mod` file.  It prints out the value and saves the go version for the other three jobs

- `lint`
  - This runs `go fmt` to make sure the code is formatted correct.  The go version is determined by the `shared-values` job.  It errors if `git diff --exit-code` fails

- `units`
  - This runs `make units` for all non-windows environments.  The go version is  determined by the `shared-values` job.  It hardcodes the version of ginkgo that is installed.   Also installs gomega matchers for some reason?

- `units-windows`
  - This runs `ginkgo` directly with some `-skipPackage` for all windows environments. The go version is determined by the `shared-values` job.  It hardcodes the version of ginkgo that is installed.   Also installs gomega matchers for some reason?

## Integration
Integration runs eighteen jobs, `shared-values`, `` and `` on pushes and pull requests to master, v8, and v7 and on manual trigger.  It will not run if the only changes are docs related.

- `shared-values`
  - This determines the go version from `go.mod` file.  The secrets environment is set to PROD.  It prints out the value and saves the go version and secrets environment for the other seventeen jobs

- `get-linux-env`
  - This claims a VMware maintained CF Deployment environment.  It uploads a metadata file with connection information for follow on jobs.  It does not specify a specific CAPI version to use.  It uses the `get-env.yml` workflow file.

- `run-linux-integration`
  - This gets the metadata file from `get-linux-env` to run integration tests and then runs the integration tests with script `integration-linux.sh`.  It uses [cloudfoundry/cf-deployment-cooncourse-tasks:latest](https://hub.docker.com/r/cloudfoundry/cf-deployment-concourse-tasks) as the container ([github link](https://github.com/cloudfoundry/cf-deployment-concourse-tasks/blob/main/dockerfiles/cf-deployment-concourse-tasks/Dockerfile))

- `unclaim-linux-env`
  - This gets the metadata file from `get-linux-env` to unclaim the VMware maintained CF Deployment environment. 


## Helper yml files
yml files use as part of integration for jobs that are often repeated

- `get-env.yml`
  - Claims a VMware maintained CF Deployment environments and deploys a specific CAPI version if requested
  - Requires a github secret `TOOLSMITHS_API_TOKEN` for access to the VMware maintained CF Deployments
  - Requires a github secret `CLIENT_SECRET` for something about bosh deploys?
  - It uses [cloudfoundry/cf-deployment-cooncourse-tasks:latest](https://hub.docker.com/r/cloudfoundry/cf-deployment-concourse-tasks) as the container ([github link](https://github.com/cloudfoundry/cf-deployment-concourse-tasks/blob/main/dockerfiles/cf-deployment-concourse-tasks/Dockerfile))

- `unclaim-env`
  - Claims the VMware maintained CF Deployment environments from a provided metadata file
  - Requires a github secret `TOOLSMITHS_API_TOKEN` for access to the VMware maintained CF Deployments
  - It uses [cloudfoundry/cf-deployment-cooncourse-tasks:latest](https://hub.docker.com/r/cloudfoundry/cf-deployment-concourse-tasks) as the container ([github link](https://github.com/cloudfoundry/cf-deployment-concourse-tasks/blob/main/dockerfiles/cf-deployment-concourse-tasks/Dockerfile))

## Scripts
See [readme](.github/workflows/scripts/readme.md) in the scripts directory