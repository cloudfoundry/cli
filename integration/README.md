# CLI Integraition Tests
These are high-level tests for the CLI that make assertions about the behavior of the `cf` binary.

These tests require that a `cf` binary built from the latest source is available in your `PATH`.

## How to run:
These tests rely on [ginkgo](https://github.com/onsi/ginkgo) to be installed. Currently there are three suites of functionality, `isolated`, `global`, and `plugin`. The `isolated` suite can run it's tests in parallel, while the `global` and `plugin` suites cannot.

Run command for the `isolated` suite:
```
ginkgo -p -r -randomizeAllSpecs -slowSpecThreshold=120 integration/isolated
```

Run command for the `global` and `plugin` suites:
```
ginkgo -r -randomizeAllSpecs -slowSpecThreshold=120 integration/isolated integration/plugin
```

### Customizations (based on environment variables):

- `CF_API` - Sets the CF API URL these tests will be using. Will default to `api.bosh-lite.com` if not set.
- `SKIP_SSL_VALIDATION` - If true, will skip SSL Validation. Will default `--skip-ssl-validation` if not set.
- `CF_USERNAME` - The CF Administrator username. Will default to `admin` if not set.
- `CF_PASSWORD` - The CF Administrator password. Will default to `admin` if not set.
- `CF_CLI_EXPERIMENTAL` - Will enable both experimental functionality of the CF CLI and tests for that functionality. Will default to `false` if not set.

### The test suite does not cleanup after itself!
In order to focus on clean test code and performance of each test, we have decided to not cleanup after each test. However, in order to facilitate [clean up scripts](https://github.com/cloudfoundry/cli/blob/master/bin/cleanup-integration), we are trying to keep consistent naming across organizations, spaces, etc.
