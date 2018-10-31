# Contributing to CLI

The Cloud Foundry team uses GitHub and accepts code contributions via [pull
requests](https://help.github.com/articles/about-pull-requests/).

## CLI V6 & V7

The code base is undergoing major work to support new APIs and there may be work
already planned that would accomplish the goals of the intended PR. The CLI team
can work with you at the start of this process to determine the best path
forward. Please contact the CLI Team prior to starting any endeavor that might
result in submitting changes back to this repository.

## Prerequisites
1. Contact the CLI team prior to working on a change. Reach out to via a GitHub
   issue or on Slack `#cli` channel at cloudfoundry.slack.com. If you're not a
   member of Cloud Foundry's slack, [request an
   invite](https://slack.cloudfoundry.org/).
1. Ensure that you have either completed our CLA Agreement for
   [individuals](https://www.cloudfoundry.org/pdfs/CFF_Individual_CLA.pdf) or
   are a [public
   member](https://help.github.com/articles/publicizing-or-hiding-organization-membership/)
   of an organization that has signed the
   [corporate](https://www.cloudfoundry.org/pdfs/CFF_Corporate_CLA.pdf) CLA.
   **The CF CLI cannot merge pull requests that contain members that have not
   signed the CFF CLA**.
1. Review the CF CLI's [Style
   Guide](https://github.com/cloudfoundry/cli/wiki/CF-CLI-Style-Guide)
1. Review the CF CLI's [Architecture
   Guide](https://github.com/cloudfoundry/cli/wiki/Architecture-Guide)
1. Review the CF CLI's [Code Style
   Guide](https://github.com/cloudfoundry/cli/wiki/Code-Style-Guide)
1. Review the CF CLI's [Testing Style
   Guide](https://github.com/cloudfoundry/cli/wiki/Testing-Style-Guide)
1. Review the CF CLI's [Internationalization
   Guide](https://github.com/cloudfoundry/cli/wiki/Internationalization-Guide)

# Development Environment Setup

## Install Golang 1.10

Documentation on installing GoLang and setting the `GOROOT`, `GOPATH` and `PATH`
environment variables can be found [here](https://golang.org/doc/install). While
the CF CLI might be compatible with other versions of GoLang, this is the only
version that the `cli` binary is built and tested with.

> To check what Golang version a particular `cf` binary was built with, use
> `strings cf | grep 'go1\....'` and look for the `go1.x.y` version number in
> the output.

## Development tools

The CF CLI requires the following development tools in order to run our test:
- [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega)
  - Test framework/Matchers Library.
- [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) - Generate
  fakes/mocks for testing. Currently using version `6.*`.
- [dep](https://github.com/golang/dep) - `vendor` dependency management tool
- [make](https://www.gnu.org/software/make/) - tool for building the CLI and
  running it's tests.

## Git Checkout

The CF CLI should **not** be checked out under `src/github.com`, instead it
should be checked out under `src/code.cloudfoundry.org`. While they resolve to
the same thing on checkout, GoLang will be unable to _correctly_ resolve them at
build time.

```bash
mkdir -p $GOPATH/src/code.cloudfoundry.org
cd $GOPATH/src/code.cloudfoundry.org
git clone https://github.com/cloudfoundry/cli.git
```

OR

```bash
go get -u github.com/cloudfoundry/cli
```

## Building the `cf` binary

Build the binary for the **current architecture** and adding it to the `PATH`:
```bash
cd $GOPATH/src/code.cloudfoundry.org/cli
make build
export PATH=$GOPATH/src/code.cloudfoundry.org/cli/out:$PATH # Puts the built CLI first in your PATH
```

### Compiling for Other Operating Systems and Architectures

The supported platforms for the CF CLI are Linux (32-bit and 64-bit), Windows
(32-bit and 64-bit) and OSX (aka Darwin). The commands that build the binaries
can be seen in the [Makefile](/Makefile) where the target begins with the
`out/cf-cli`.


For general information on how to cross compile GoLang binaries, see the [Go
environment variables
documentation](https://golang.org/doc/install/source#environment) for details on
how to cross compile binaries for other architectures.

## Running the Unit tests

To run the unit tests:
```bash
cd $GOPATH/src/code.cloudfoundry.org/cli
make units-full # will run all unit tests
make units # runs all non-cf directory unit tests
```

**Note: `make units-full` is recommended over `make units` if you are unsure of
how wide reaching the intended changes are.**

## Running the Integration tests

The [Integration test README](/integration/README.md) contains a full set of
details on how to configure and run the integration tests. In addition to the
configuration mentioned in the README, the CLI's `Makefile` contains the
following support commands that will run `make build cleanup-integration` prior
to running integration tests:

```bash
make integration-experimental # runs the experimental integration tests
make integration-global # runs the global integration tests
make integration-isolated # runs the isolated integration tests
make integration-plugin # runs the plugin integration tests
make integration-push # runs the push integration tests
make integration-tests # runs the isolated, push and global integration tests
make integration-tests-full # runs all the integration suites
```

If the number of parallel nodes for the non-global test suites would like to be
adjusted, set the `NODES` environment variable:

```bash
NODES=10 make integration-tests
```

# Modifying the CLI codebase

All changes to the CF CLI require updates to the unit/integration test. There
are additional requirements around updating the CF CLI that will be listed
below.

## Updating counterfeiter fakes

The CLI uses [`counterfeiter`](https://github.com/maxbrunsfeld/counterfeiter) to
generate fakes from interfaces for the unit tests. If any changes are made to an
interface, the fakes be should regenerated using counterfeiter:

```bash
go generate ./<package>/...
```

where `<package>` contains the package with the changed interface.

### Notes
1. `counterfeiter` fakes should never be manually edited. They are only
   created/modified via `go generate`. **All pull requests with manually modified
   fakes will be rejected.**
1. Do not run `go generate` from the root directory. Fakes in the legacy
   codebase require additional intervention so it preferred not to modify them
   unless it is _absolutely_ necessary.

## Vendoring Dependencies

The CLI uses [`dep`](https://github.com/golang/dep) to manage vendored
dependencies. Refer to the [`dep`
documentation](https://golang.github.io/dep/docs/daily-dep.html) for managing
dependencies.

If you are vendoring a new dependency, please read [License and Notice
Files](https://github.com/cloudfoundry/cli/wiki/License-and-Notice-Files) to
abide by third party licenses.

## API Versioning

The CLI has a minimum version requirements for the APIs it interfaces with, the
requirements for these APIs are listed in the [Version Policy
guide](https://github.com/cloudfoundry/cli/wiki/Versioning-Policy#cf-cli-minimum-supported-version).

If your pull request requires a CAPI version higher than the minimum API version,
the CLI code and integration tests must be versioned tests. This new
functionality has the following requirements:

1. The minimum version is added to the [Minimum API version
   list](/api/cloudcontroller/ccversion/minimum_version.go).
1. The feature has an appropriate version check in the `command` layer to prevent
   use of that feature if the targeted API is below the minimum version. **Note:
   commands should FAIL prior to execution when minimum version is not met for
   specified functionality.**
1. The integration tests that are added use the `helpers.SkipIfVersionLessThan`
   or `helpers.SkipIfVersionGreaterThan` helpers in their `BeforeEach`. See this
   [example](https://github.com/cloudfoundry/cli/blob/87aaed8215fad3b2077c6829d1812ead3902d5cf/integration/isolated/create_isolation_segment_command_test.go#L17).

## i18n translations

The CF CLI has internal infrastructure in place to support multiple languages
for UI output. This means that all text outputted to the user must pass through
the internal translation layer. For more information on how this works, read the
[Internationalization
Guide](https://github.com/cloudfoundry/cli/wiki/Internationalization-Guide).
