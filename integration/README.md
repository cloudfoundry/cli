# CLI Integration Tests

## Introduction

These are high-level tests for the CLI that make assertions about the behavior of the `cf` binary.

On most systems `cf` points to an installed version. To test the latest source (most likely source that you're changing), ensure the dev `cf` binary is in your `PATH`:

```
[[ `which cf` = *"$GOPATH/src/code.cloudfoundry.org/cli/out"* ]] || 
    export PATH="$GOPATH/src/code.cloudfoundry.org/cli/out:$PATH"
```

You'll also need to rebuild `cf` after making any relevant changes to the source:

```
make build
```

Running `make integration-tests` can be time-consuming, because it includes the unparallelized `global` suite. Best to constrain runs to relevant tests until a long break in your workday, when you can run `make integration-tests` and cover everything. If you're primarily working in code that is tested by the parallelized suites, running the rake tasks for those specific suites instead of `integration-tests` and setting the `NODES` environment variable to a higher value will improve your feedback cycle.

## Explanation of test suites
- `global` suite is for tests that affect an entire CF instance. *These tests do not run in parallel.*
- `isolated` suite is for tests that are stand alone and do not affect each other. They are meant to run in their own organization and space, and will not affect system state. This is the most common type of integration tests.
- `push` suite is for tests related to the `cf push` command only.
- `experimental` suite is for tests that require the cf experimental flag to be set and/or an experimental feature for the CF CLI.
- `plugin` suite is for tests that surround the CF CLI plugin framework. *These tests do not run in parallel.*

## How to run
These tests rely on [ginkgo](https://github.com/onsi/ginkgo) to be installed.

Run command for the `isolated`, `push` and `experimental` suite:
```
ginkgo -p -r -randomizeAllSpecs -slowSpecThreshold=120 integration/shared/isolated integration/v6/push integration/shared/experimental
```

Run command for the `global` and `plugin` suites:
```
ginkgo -r -randomizeAllSpecs -slowSpecThreshold=120 integration/shared/global integration/shared/plugin
```

### Customizations (based on environment variables)

- `CF_INT_API` - Sets the CF API URL these tests will be using. Will default to `api.bosh-lite.com` if not set.
- `SKIP_SSL_VALIDATION` - If true, will skip SSL Validation. Will default `--skip-ssl-validation` if not set.
- `CF_INT_USERNAME` - The CF Administrator username. Will default to `admin` if not set.
- `CF_INT_PASSWORD` - The CF Administrator password. Will default to `admin` if not set.
- `CF_INT_OIDC_USERNAME` - The admin user in the OIDC identity provider. Will default to `admin_oidc` if not set.
- `CF_INT_OIDC_PASSWORD` - The admin password in the OIDC identity provider. Will default to `admin` if not set.
- `CF_INT_DOCKER_IMAGE` - A private docker image used for the docker authentication tests.
- `CF_INT_DOCKER_USERNAME` - The username for the private docker registry for `CF_INT_DOCKER_IMAGE`.
- `CF_INT_DOCKER_PASSWORD` - The password for `CF_INT_DOCKER_USERNAME`.
- `CF_INT_CLIENT_ID` - the ID for the integration client credentials identity.
- `CF_INT_CLIENT_SECRET` - the secret for the integration client credentials identity.
- `CF_CLI_EXPERIMENTAL` - Will enable both experimental functionality of the CF CLI and tests for that functionality. Will default to `false` if not set.

### The test suite does not cleanup after itself!
In order to focus on clean test code and performance of each test, we have decided to not cleanup after each test. However, in order to facilitate [clean up scripts](https://github.com/cloudfoundry/cli/blob/master/bin/cleanup-integration), we are trying to keep consistent naming across organizations, spaces, etc.

In addition, several router groups are created using a `INTEGRATION-TCP-NODE-[NUMBER]` format. These cannot be deleted without manual changes to the database.
