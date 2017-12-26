# CLI Integration Tests
These are high-level tests for the CLI that make assertions about the behavior of the `cf` binary.

These tests require that a `cf` binary built from the latest source is available in your `PATH`.

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
ginkgo -p -r -randomizeAllSpecs -slowSpecThreshold=120 integration/isolated integration/push integration/experimental
```

Run command for the `global` and `plugin` suites:
```
ginkgo -r -randomizeAllSpecs -slowSpecThreshold=120 integration/global integration/plugin
```

### Customizations (based on environment variables)

- `CF_API` - Sets the CF API URL these tests will be using. Will default to `api.bosh-lite.com` if not set.
- `SKIP_SSL_VALIDATION` - If true, will skip SSL Validation. Will default `--skip-ssl-validation` if not set.
- `CF_USERNAME` - The CF Administrator username. Will default to `admin` if not set.
- `CF_PASSWORD` - The CF Administrator password. Will default to `admin` if not set.
- `CF_INT_DOCKER_IMAGE` - A private docker image used for the docker authentication tests.
- `CF_INT_DOCKER_USERNAME` - The username for the private docker registry for `CF_INT_DOCKER_IMAGE`.
- `CF_INT_DOCKER_PASSWORD` - The password for `CF_INT_DOCKER_USERNAME`.
- `CF_CLI_EXPERIMENTAL` - Will enable both experimental functionality of the CF CLI and tests for that functionality. Will default to `false` if not set.

### The test suite does not cleanup after itself!
In order to focus on clean test code and performance of each test, we have decided to not cleanup after each test. However, in order to facilitate [clean up scripts](https://github.com/cloudfoundry/cli/blob/master/bin/cleanup-integration), we are trying to keep consistent naming across organizations, spaces, etc.

In addition, several router groups are created using a `INTEGRATION-TCP-NODE-[NUMBER]` format. These cannot be deleted without manual changes to the database.
