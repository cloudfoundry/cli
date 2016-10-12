CLI Acceptance Tests
====
These are high-level tests for the [Cloud Foundry
CLI](https://github.com/cloudfoundry/cli) that make assertions about the
behavior of the `cf` binary.

These tests require that a `cf` binary built from the latest source is
available in your `PATH`.

# New Suite:
These are the notes for the `integration` suite tests.

# How to run:
Running is simple:

```
ginkgo -p -r -randomizeAllSpecs -slowSpecThreshold=120 integration
```

Customizations (based on environment variables):

- `CF_API` - Sets the CF API URL these tests will be using. Will default to `api.bosh-lite.com` if not set.
- `SKIP_SSL_VALIDATION` - If true, will skip SSL Validation. Will default `--skip-ssl-validation` if not set.
