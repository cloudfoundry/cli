# Development Environment Setup

## Install Golang 1.7.1 or higher

Documentation on installing Golang and setting the `GOROOT`, `GOPATH` and `PATH` environment variables can be found [here](https://golang.org/doc/install).

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

# Contributing to CLI

The Cloud Foundry team uses GitHub and accepts code contributions via
[pull requests](https://help.github.com/articles/using-pull-requests).
If your contribution includes a change that is exposed to cf CLI users
(e.g. introducing a new command or flag), please submit an issue
to discuss it first.
Major new feature proposals generally entail a publicly viewable
google document with commenting allowed to be discussed on the [cf-dev](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/) mailing list.

## Contributor License Agreement

Follow these steps to make a contribution to any of our open source repositories:

1. Ensure that you have completed our CLA Agreement for
  [individuals](https://www.cloudfoundry.org/wp-content/uploads/2015/09/CFF_Individual_CLA.pdf) or
  [corporations](https://www.cloudfoundry.org/wp-content/uploads/2015/09/CFF_Corporate_CLA.pdf).
