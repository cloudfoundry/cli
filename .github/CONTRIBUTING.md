# Contributing to CLI

The Cloud Foundry team uses GitHub and accepts code contributions via
[pull requests](https://help.github.com/articles/using-pull-requests).

## CLI v6.x & v7-beta

The code base is undergoing major work to support new Cloud Controller APIs.
Most CLI users have a v6 CLI.
Some users are trying out a beta version of v7.
Both versions are in active development.
If you are not sure, you probably want to modify the v6 CLI.
The CLI team can work with you at the start of this process 
to determine the best path forward.
You will find the entry point for v6 commands in the _cli/command/v6_ directory.
v7 commands are found in the _cli/command/v7_ directory.
More details are available in the [Architecture Guide](https://github.com/cloudfoundry/cli/wiki/Architecture-Guide).
  
## Prerequisites

Before working on a PR to the CLI code base, please:

  - chat with us on our [Slack #cli channel](https://cloudfoundry.slack.com) ([request an invite](http://slack.cloudfoundry.org/)),
  - reach out to us first via a [GitHub issue](https://github.com/cloudfoundry/cli/issues),
  - or, look for the [contributions welcome label on GitHub](https://github.com/cloudfoundry/cli/issues?q=is%3Aopen+is%3Aissue+label%3A%22contributions+welcome%22) 
    for issues we are actively looking for help on.

After reaching out to the CLI team and the conclusion is to make a PR, please follow these steps:

1. Ensure that you have either 
   * completed our [Contributor License Agreement (CLA) for individuals](https://www.cloudfoundry.org/pdfs/CFF_Individual_CLA.pdf), 
   * or, are a [public member](https://help.github.com/articles/publicizing-or-hiding-organization-membership/) of an organization 
   that has signed [the corporate CLA](https://www.cloudfoundry.org/pdfs/CFF_Corporate_CLA.pdf).
1. Review the CF CLI [Style Guide](https://github.com/cloudfoundry/cli/wiki/CF-CLI-Style-Guide).
   Feel free to peruse our
   [Architecture Guide](https://github.com/cloudfoundry/cli/wiki/Architecture-Guide),
   [Code Style Guide](https://github.com/cloudfoundry/cli/wiki/Code-Style-Guide),
   [Testing Style Guide](https://github.com/cloudfoundry/cli/wiki/Testing-Style-Guide),
   or [Internationalization Guide](https://github.com/cloudfoundry/cli/wiki/Internationalization-Guide).
1. Fork the project repository
1. Create a feature branch (e.g. `git checkout -b better_cli`) and make changes on this branch
   * Follow the [other sections on this page](#development-environment-setup) to set up your development environment, build `cf` and run the tests.
   * Tests are required for any changes.
1. Push to your fork (e.g. `git push origin better_cli`) and [submit a pull request](https://help.github.com/articles/creating-a-pull-request)

Note: All contributions must be sent using GitHub Pull Requests.
We prefer a small, focused pull request with a clear message 
that conveys the intent of your change. 
Please make sure to squash commits into meaningful chunks of work. 

# Development Environment Setup

## Install Golang 1.10

We defer to [the official Golang documentation](https://golang.org/doc/install) for installing Golang 
and setting the `GOROOT`, `GOPATH` and `PATH` environment variables.
While the CF CLI might be compatible with other versions, 
Golang 1.10 is the only version that the `cli` binary is built and tested with.

> To check what Golang version a particular Linux `cf` binary was built with, use `strings cf | grep 'go1\....'` and look for the `go1.x.y` version number in the output.


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


## Install bosh-lite and deploy Cloud Foundry

The CLI integration tests require a Cloud Foundry deployment. The easiest way to
deploy a local Cloud Foundry for testing is to use bosh-lite - see
[Quick Start](https://bosh.io/docs/quick-start/) and [cf-deployment](https://github.com/cloudfoundry/cf-deployment) to deploy CF.

The `ci/local-integration-env` folder contains scripts and documentation for
deploying a Cloud Foundry with the configuration that our integration tests
expect to be present.

Before using the scripts, make sure that you have installed the following
dependencies:
* BOSH CLI and its
  [dependencies](https://bosh.io/docs/cli-v2-install/#additional-dependencies)
* [VirtualBox](https://www.virtualbox.org/)

## Updating tests

The CLI uses [`counterfeiter`](https://github.com/maxbrunsfeld/counterfeiter) to generate unit test fakes from interfaces. If you make any changes to an interface you should regenerate the fakes by running:
```
go get -u github.com/maxbrunsfeld/counterfeiter # install counterfeiter

go generate ./<directory>/...
```
where `<directory>` contains the package with the changed interface. Don't run `go generate` from the root directory.

The CLI has a minimum required version. Refer to the following:

https://github.com/cloudfoundry/cli/wiki/Versioning-Policy#cf-cli-minimum-supported-version

If your pull request requires a CAPI version higher than the minimum, integration tests you implement must be versioned tests. To do so please add your minimum version to `api/cloudcontroller/ccversion/minimum_version.go`, and a corresponding `helpers.SkipIfVersionLessThan` or `helpers.SkipIfVersionGreaterThan`. See this [example](https://github.com/cloudfoundry/cli/blob/87aaed8215fad3b2077c6829d1812ead3902d5cf/integration/isolated/create_isolation_segment_command_test.go#L17).

## Running tests

First install `ginkgo`.
```
go get -u github.com/onsi/ginkgo/ginkgo
```

### Running unit tests

Run the tests:
```
cd $GOPATH/src/code.cloudfoundry.org/cli

make test
```

### Running integration tests

If you have a BOSH-lite Cloud Foundry running as described in [the above section](https://github.com/cloudfoundry/cli/blob/master/.github/CONTRIBUTING.md#install-bosh-lite-and-deploy-cloud-foundry), all you need to do is run:
```
cd $GOPATH/src/code.cloudfoundry.org/cli

make integration-tests
```

If you want to target a different Cloud Foundry, set the following environment
variables before running tests:
```
export CF_INT_API=api.my-cf-domain.com
export CF_INT_PASSWORD=my-admin-cf-password
```

More information on the integration tests, such as descriptions of the suites
and integration environment variables, can be found in the [integration
README](https://github.com/cloudfoundry/cli/blob/master/integration/README.md)

# Architecture Overview

The CLI is divided into a few major components, including but not limited to:

1. command
1. actor
1. API

#### command
The command package is the gateway to each CLI command accessible to the CLI, using the actors to talk to the API. Each command on the CLI has 1 corresponding file in the command package. The command package is also responsible for displaying the UI.

#### actor
The actor package consists of one actor that handles all the logic to process the commands in the CLI. Actor functions are shared workflows that can be used by more than one command. The functions may call upon several API calls to implement their business logic.

#### API
The API package handles the HTTP requests to the API. The functions in this package return a resource that the actor can then parse and handle. The structures returned by this package closely resemble the return bodies of the Cloud Controller API.

For more information, check out our [Architecture Guide](https://github.com/cloudfoundry/cli/wiki/Architecture-Guide)

# Vendoring Dependencies

The CLI uses [dep](https://github.com/golang/dep) to manage vendored
dependencies. Refer to the [`dep` documentation](https://golang.github.io/dep/docs/daily-dep.html) for managing dependencies.

If you are vendoring a new dependency, please read [License and Notice Files](https://github.com/cloudfoundry/cli/wiki/License-and-Notice-Files) to abide by third party licenses.

# Compiling for Other Operating Systems and Architectures

The supported platforms for the CF CLI are Linux (32-bit and 64-bit), Windows
(32-bit and 64-bit) and macOS. The commands that build the binaries can be seen
in the [build binaries Concourse task](https://github.com/cloudfoundry/cli/blob/master/ci/cli/tasks/build-binaries.yml).

See the [Go environment variables documentation](https://golang.org/doc/install/source#environment)
for details on how to cross compile binaries for other architectures.

# i18n translations

If you are adding new strings or updating existing strings within the CLI code, you'll need to update the binary representation of the translation files. This file is generated/maintained using [i18n4go](https://github.com/XenoPhex/i18n4go), [goi18n](https://github.com/nicksnyder/go-i18n), and `bin/generate-language-resources`.

After adding/changing strings supplied to the goi18n `T()` translation func, run the following to update the translations binary:

    i18n4go -c fixup # answer any prompts appropriately
    goi18n -outdir cf/i18n/resources cf/i18n/resources/*.all.json
    bin/generate-language-resources

When running `i18n4go -c fixup`, you will be presented with the choices `new` or `upd` for each addition or update. Type in the appropriate choice. If `upd` is chosen, you will be asked to confirm which string is being updated using a numbered list.

After running the above, be sure to commit the translations binary, `cf/resources/i18n_resources.go`.

# Plugins

* [CF CLI plugin development guide](https://github.com/cloudfoundry/cli/tree/master/plugin/plugin_examples)
* [plugins repository](https://plugins.cloudfoundry.org/)

When importing the plugin code use `import "code.cloudfoundry.org/cli/plugin"`.
Older plugins that import `github.com/cloudfoundry/cli/plugin` will still work
as long they vendor the plugins directory.
