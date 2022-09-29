# Scripts
Scripts used to run integration tests in github

## `integration-linux.sh`
- Parses the metafile for environment connection details and sets environment variables.  Installs specific ginkgo version and runs `make integration-tests-full-ci`