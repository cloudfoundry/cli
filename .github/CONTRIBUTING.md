# Contributing to CLI

The Cloud Foundry team uses GitHub and accepts code contributions via [pull
requests](https://help.github.com/articles/about-pull-requests/).

## CLI v6.x & v7-beta

The code base is undergoing major work to support new Cloud Controller APIs.
Most CLI users have a v6 CLI.
Some users are trying out a beta version of v7.
Both versions are in active development.
You will find the entry point for v6 commands in the _cli/command/v6_ directory.
v7 commands are found in the _cli/command/v7_ directory.
More details are available in the [Architecture Guide](https://github.com/cloudfoundry/cli/wiki/Architecture-Guide).

**Note**: The CLI can currently be compiled to use V6 or V7 code paths.
Depending on the nature of the intended changes, you may be contributing to V6
code, V7 code, and/or shared code. The rest of this guide assumes that
your changes are to V6 code or shared code. If your changes are in V7 code, refer to the
[V7-specific contributing information](https://github.com/cloudfoundry/cli/wiki/Contributing-V7.md) for more information.
If your changes are to shared code, please also run V7 tests before submitting a
pull request.

The `make` commands in this file can be used with V6 or V7 code. To run the V7
version of tests or build the V7 version of the binary, set `TARGET_V7=1` in the
environment. If `TARGET_V7` is unset, `make` commands will target V6.

## Prerequisites

Before working on a PR to the CLI code base, please:

  - chat with us on our [Slack #cli channel](https://cloudfoundry.slack.com) ([request an invite](http://slack.cloudfoundry.org/)),
  - reach out to us first via a [GitHub issue](https://github.com/cloudfoundry/cli/issues),
  - or, look for the [contributions welcome label on GitHub](https://github.com/cloudfoundry/cli/issues?q=is%3Aopen+is%3Aissue+label%3A%22contributions+welcome%22)
    for issues we are actively looking for help on.

After reaching out to the CLI team and the conclusion is to make a PR, please follow these steps:

1. Ensure that you have either:
   * completed our [Contributor License Agreement (CLA) for individuals](https://www.cloudfoundry.org/pdfs/CFF_Individual_CLA.pdf),
   * or, are a [public member](https://help.github.com/articles/publicizing-or-hiding-organization-membership/) of an organization
   that has signed [the corporate CLA](https://www.cloudfoundry.org/pdfs/CFF_Corporate_CLA.pdf).
1. Review the CF CLI [Style Guide](https://github.com/cloudfoundry/cli/wiki/CF-CLI-Style-Guide),
   [Architecture Guide](https://github.com/cloudfoundry/cli/wiki/Architecture-Guide),
   [Code Style Guide](https://github.com/cloudfoundry/cli/wiki/Code-Style-Guide),
   [Testing Style Guide](https://github.com/cloudfoundry/cli/wiki/Testing-Style-Guide),
   or [Internationalization Guide](https://github.com/cloudfoundry/cli/wiki/Internationalization-Guide).
1. Fork the project repository.
1. Create a feature branch (e.g. `git checkout -b better_cli`) and make changes on this branch
   * Follow the other sections on this page to [set up your development environment](#development-environment-setup), [build `cf`](#building-the-cf-binary) and [run the tests](#testing).
   * Tests are required for any changes.
1. Push to your fork (e.g. `git push origin better_cli`) and [submit a pull request](https://help.github.com/articles/creating-a-pull-request)

Note: All contributions must be sent using GitHub Pull Requests.
We prefer a small, focused pull request with a clear message
that conveys the intent of your change.

# Development Environment Setup

## Install Golang 1.11

Documentation on installing GoLang can be found [here](https://golang.org/doc/install). While
the CF CLI might be compatible with other versions of GoLang, this is the only
version that the `cli` binary is built and tested with.

## Development tools

The CF CLI requires the following development tools in order to run our test:
- [Ginkgo](https://github.com/onsi/ginkgo)/[Gomega](https://github.com/onsi/gomega) - Test framework/Matchers Library.
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

# Building the `cf` binary

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

# Testing

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
