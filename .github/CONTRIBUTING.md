# Contributing to CLI

The Cloud Foundry team uses GitHub and accepts code contributions via
[pull requests](https://help.github.com/articles/using-pull-requests).

Before working on a PR to the CLI code base, please reach out to us first via a GitHub issue or on our Slack #cli channel at cloudfoundry.slack.com. (If you're not a member of Cloud Foundry's slack team yet, you'll need to [request an invite](https://slack.cloudfoundry.org/).) There are areas of the code base that contain a lot of complexity, and something which seems like a simple change may be more involved. In addition, the code base is undergoing re-architecturing/refactoring, and there may be work already planned that would accomplish the goals of the intended PR. The CLI team can work with you at the start of this process to determine the best path forward.

After reaching out to the CLI team and the conclusion is to make a PR, please follow these steps:
1. Ensure that you have either completed our CLA Agreement for [individuals](https://www.cloudfoundry.org/pdfs/CFF_Individual_CLA.pdf) or are a [public member](https://help.github.com/articles/publicizing-or-hiding-organization-membership/) of an organization that has signed the [corporate](https://www.cloudfoundry.org/pdfs/CFF_Corporate_CLA.pdf) CLA.
1. Fork the project’s repository
1. Create a feature branch (e.g. `git checkout -b better_cli`) and make changes on this branch
  * Follow the previous sections on this page to set up your development environment, build `cf` and run the tests.
1. Push to your fork (e.g. `git push origin better_cli`) and [submit a pull request](https://help.github.com/articles/creating-a-pull-request)

If you have a CLA on file, your contribution will be analyzed for product fit and engineering quality prior to merging.  
Note: All contributions must be sent using GitHub Pull Requests.  
**Your pull request is much more likely to be accepted if it is small and focused with a clear message that conveys the intent of your change. Please make sure to squash commits into meaningful chunks of work.** Tests are required for any changes.

# Development Environment Setup

## Install Golang 1.9 or higher

Documentation on installing GoLang and setting the `GOROOT`, `GOPATH` and `PATH` environment variables can be found [here](https://golang.org/doc/install). While the CF CLI might be compatible with other versions of GoLang, this is the only version that the `cli` binary is built and tested with.

> To check what Golang version a particular Linux `cf` binary was built with, use `strings cf | grep 'go1\....'` and look for the `go1.x.y` version number in the output.

## Obtain the Source

```sh
go get code.cloudfoundry.org/cli
```

## Building the `cf` binary

Build the binary and add it to the PATH:
```
cd $GOPATH/src/code.cloudfoundry.org/cli
bin/build
export PATH=$GOPATH/src/code.cloudfoundry.org/cli/out:$PATH
```

## Install bosh-lite and deploy Cloud Foundry

The CLI integration tests need a Cloud Foundry deployment. The easiest way to
deploy a local Cloud Foundry for testing is to use
[bosh-lite](https://github.com/cloudfoundry/bosh-lite). Follow these
instructions:

https://github.com/cloudfoundry/bosh-lite#deploy-cloud-foundry

## Run the Tests

First install `ginkgo`.
```
go get -u github.com/onsi/ginkgo/ginkgo
```

Run the tests:
```
cd $GOPATH/src/code.cloudfoundry.org/cli

ginkgo -r
```

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

# Vendoring Dependencies

The CLI uses [GVT](https://github.com/FiloSottile/gvt) to manage vendored
dependencies. Refer to the GVT documentation for managing dependencies.

If you are vendoring a new dependency, please read [License and Notice Files](https://github.com/cloudfoundry/cli/wiki/License-and-Notice-Files) to abide by third party licenses.

# Compiling for Other Operating Systems and Architectures

The supported platforms for the CF CLI are Linux (32-bit and 64-bit), Windows
(32-bit and 64-bit) and OSX. The commands that build the binaries can be seen
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
