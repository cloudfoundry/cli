# Contributing to CLI

The Cloud Foundry team uses GitHub and accepts code contributions via [pull
requests](https://help.github.com/articles/about-pull-requests/). If you have any questions, ask away on the #cli channel in [our Slack
community](https://slack.cloudfoundry.org/) and the
[cf-dev](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/)
mailing list.

## CLI Versions
The cf CLI follows a branching model:
- V9 (Next major release) of the cf CLI is built from the [main branch](https://github.com/cloudfoundry/cli/tree/main). **This branch is under active development**.
- V8 of the cf CLI is built from the [v8 branch](https://github.com/cloudfoundry/cli/tree/v8). **This branch is under active development**.
- V7 of the cf CLI is built from the [v7 branch](https://github.com/cloudfoundry/cli/tree/v7). **This branch is no longer maintained or released**.
- V6 of the cf CLI is built from the [v6 branch](https://github.com/cloudfoundry/cli/tree/v6). **This branch is no longer maintained or released**.

## Prerequisites
Before working on a PR to the CLI code base, please:

  - reach out to us first via a [GitHub issue](https://github.com/cloudfoundry/cli/issues),
  - look for the [contributions welcome label on GitHub](https://github.com/cloudfoundry/cli/issues?q=is%3Aopen+is%3Aissue+label%3A%22contributions+welcome%22)
    for issues we are actively looking for help on.

You can always chat with us on our [Slack #cli channel](https://cloudfoundry.slack.com) ([request an invite](http://slack.cloudfoundry.org/)),

After reaching out to the CLI team and the conclusion is to make a PR, please follow these steps:

1. Ensure that you have either:
   * completed our [Contributor License Agreement (CLA) for individuals](https://www.cloudfoundry.org/pdfs/CFF_Individual_CLA.pdf),
   * or, are a [public member](https://help.github.com/articles/publicizing-or-hiding-organization-membership/) of an organization
   that has signed [the corporate CLA](https://www.cloudfoundry.org/pdfs/CFF_Corporate_CLA.pdf).
1. Review the CF CLI [Style Guide](https://github.com/cloudfoundry/cli/wiki/CF-CLI-Style-Guide),
   [Architecture Guide](https://github.com/cloudfoundry/cli/wiki/Architecture-Guide),
   [Product Style Guide](https://github.com/cloudfoundry/cli/wiki/CLI-Product-Specific-Style-Guide),
   and [Internationalization Guide](https://github.com/cloudfoundry/cli/wiki/Internationalization-Guide).
1. Fork the project repository.
1. Create a feature branch from the earliest branch that's [appropriate for your change](#cli-versions) (e.g. `git checkout v8 && git checkout -b better_cli`) and make changes on this branch
   * Follow the other sections on this page to [set up your development environment](#development-environment-setup), [build `cf`](#building-the-cf-binary) and [run the tests](#testing).
   * Tests are required for any changes.
1. Push to your fork (e.g. `git push origin better_cli`) and [submit a pull request](https://help.github.com/articles/creating-a-pull-request)
1. The cf CLI team will merge your changes from the versioned branch (e.g. v8) to main for you after the PR is merged.

Note: All contributions must be sent using GitHub Pull Requests.
We prefer a small, focused pull request with a clear message
that conveys the intent of your change.

# Development Environment Setup

## Install Golang 1.18

Documentation on installing GoLang can be found [here](https://golang.org/doc/install). While
the CF CLI might be compatible with other versions of GoLang, this is the only
version that the `cli` binary is built and tested with.

## Development tools

The CF CLI requires the following development tools in order to run our test:
- [Ginkgo](https://github.com/onsi/ginkgo) / [Gomega](https://github.com/onsi/gomega) - Test framework/Matchers Library
- [golangci-lint](https://github.com/golangci/golangci-lint) - Comprehensive linting tool
- [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) - Generate
  fakes/mocks for testing. Currently using version `6.*`.
- [make](https://www.gnu.org/software/make/) - tool for building the CLI and
  running its tests.

## Git Checkout

Clone the repository.
```bash
git clone https://github.com/cloudfoundry/cli.git
```

# Building the `cf` binary

Build the binary for the **current architecture** and adding it to the `PATH`:
```bash
cd cli
make build
export PATH=<path-to-cli-directory>/out:$PATH # Puts the built CLI first in your PATH
```

### Compiling for Other Operating Systems and Architectures

The supported platforms for the CF CLI are Linux (x86, x86-64 and arm64) , Windows
(x86 and x86-64) and OSX (aka Darwin x86-64 and arm64). The commands that build the binaries
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
cd cli
make units-full # will run all unit tests
make units # runs all non-cf directory unit tests
```

**Note: `make units-full` is recommended over `make units` if you are unsure of
how wide-reaching the intended changes are.**

## Running the Integration tests

The [Integration test README](/integration/README.md) contains a full set of
details on how to configure and run the integration tests. In addition to the
configuration mentioned in the README, the CLI's `Makefile` contains the
following support commands that will run `make build integration-cleanup` prior
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

To adjust the number of parallel nodes for the non-global test suites, set the
`NODES` environment variable:

```bash
NODES=10 make integration-tests
```

# Modifying the CLI codebase

All changes to the CF CLI require updates to the unit/integration tests. There
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

The CLI uses [`go modules`](https://golang.org/ref/mod) to manage
dependencies. Refer to the [`vendoring section`
documentation](https://golang.org/ref/mod#vendoring) for managing
dependencies.

If you are vendoring a new dependency, please read [License and Notice
Files](https://github.com/cloudfoundry/cli/wiki/License-and-Notice-Files) to
abide by third party licenses.

## API Versioning

The CLI has a minimum version requirements for the APIs it interfaces with. The
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
