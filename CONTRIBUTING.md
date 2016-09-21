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

## Setup

1. Install [Go 1.7.1](https://golang.org/dl/) or up
1. Create a directory where you would like to store the source for Go projects and their binaries (e.g. `$HOME/go`)
1. Set an environment variable, `GOPATH`, pointing at the directory you created
1. Get the `cf` source: `go get github.com/cloudfoundry/cli`
  * (Ignore any warnings about "no buildable Go source files")
1. [Fork this repository](https://help.github.com/articles/fork-a-repo/), adding your fork as a remote
1. Run our bootstrap script, `bin/bootstrap`

## Compiling the Binary

This will build a static binary, without ``-tags netgo``, it will dynamically link to the local networking library.

### Linux 32-bit

```
CGO_ENABLED=0 GOARCH=386 GOOS=linux go build -a -tags netgo -installsuffix netgo -ldflags '-extldflags "-static"' -o out/cf-cli_linux_i686 .
```

### Linux 64-bit

```
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -tags netgo -installsuffix netgo -ldflags '-extldflags "-static"' -o out/cf-cli_linux_x86-64 .
```

### Windows 32-bit

```
GOARCH=386 GOOS=windows go build -tags="forceposix" -o out/cf-cli_win32.exe .
```

### Windows 64-bit

```
GOARCH=amd64 GOOS=windows go build -tags="forceposix" -o out/cf-cli_winx64.exe .
```

### OSX

```
GOARCH=amd64 GOOS=darwin go build -o out/cf-cli_osx .
```

## Workflow

1. Run all the existing tests with `bin/test` to ensure they pass
1. Write a new test, see it fail when running `bin/test` (or `ginkgo -p path/to/the/package/being/tested`)
1. Write code to pass the test
1. Repeat the above two steps until the feature is complete
1. Run all the existing tests with `bin/test` to ensure they *still* pass
1. Submit a [pull request](https://help.github.com/articles/using-pull-requests/) to the `master` branch

**_*_ For development guide on writing a cli plugin, see [here](https://github.com/cloudfoundry/cli/tree/master/plugin_examples)**

## Architecture Overview

A command is a struct that implements this interface:

```Go
type Command interface {
	MetaData() CommandMetadata
	SetDependency(deps Dependency, pluginCall bool) Command
	Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) []requirements.Requirement
	Execute(context flags.FlagContext)
}
```
[Source code](https://github.com/cloudfoundry/cli/blob/master/cf/commandregistry/command.go#L9)

`Metadata()` is just a description of the command name, usage and flags:
```Go
type CommandMetadata struct {
	Name            string
	ShortName       string
	Usage           []string
	Description     string
	Flags           map[string]flags.FlagSet
	SkipFlagParsing bool
	TotalArgs       int
	Examples        []string
}
```
[Source code](https://github.com/cloudfoundry/cli/blob/master/cf/commandregistry/command.go#L16)

The `Examples` field represents the set of lines to be printed when printing examples in the help text.

`Requirements()` returns a list of requirements that need to be met before a command can be invoked.

`Execute()` is the method that your command implements to do whatever it's supposed to do. The `context` object
provides flags and arguments.

When the command is run, it communicates with api using repositories (they are in [`cf/api`](https://github.com/cloudfoundry/cli/blob/master/cf/api)).

`SetDependency()` is where a command obtains its dependencies. Dependencies are typically declared as an interface type, and not a concrete type, so tests can inject a fake.
The bool argument `pluginCall` indicates whether the command is invoked by one of the CLI's plugin API methods.

Dependencies are injected into each command, so tests can inject a fake. This means that dependencies are
typically declared as an interface type, and not a concrete type. (see [`cf/commandregistry/dependency.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commandregistry/dependency.go))

Some dependencies are managed by a repository locator in [`cf/api/repository_locator.go`](https://github.com/cloudfoundry/cli/blob/master/cf/api/repository_locator.go).

Repositories communicate with the api endpoints through a Gateway (see [`cf/net`](https://github.com/cloudfoundry/cli/tree/master/cf/net)).

Models are data structures related to Cloud Foundry (see [`cf/models`](https://github.com/cloudfoundry/cli/tree/master/cf/models)). For example, some models are
apps, buildpacks, domains, etc.

## Managing Dependencies

Command dependencies are managed by the command registry package. The app uses the package (in [`cf/commandregistry/dependency.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commandregistry/dependency.go)) to instantiate them, this allows not sharing the knowledge of their dependencies with the app itself.

For commands that use another command as dependency, `commandregistry` is used for retrieving the command dependency. For example, the command `restart` has a dependency on command `start` and `stop`, and this is how the command dependency is retrieved: [`restart.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commands/application/restart.go#L59)

As for repositories, we use the repository locator to handle their dependencies. You can find it in [`cf/api/repository_locator.go`](https://github.com/cloudfoundry/cli/blob/master/cf/api/repository_locator.go).

## Example Command

Create Space is a good example of a command. Its tests include checking arguments, requiring the user
to be logged in, and the actual behavior of the command itself. You can find it in [`cf/commands/space/create_space.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commands/space/create_space.go).

## i18n

If you are adding new strings or updating existing strings within the CLI code, you'll need to update the binary representation of the translation files. This file is generated/maintained using [i18n4go](https://github.com/XenoPhex/i18n4go), [goi18n](https://github.com/nicksnyder/go-i18n), and `bin/generate-language-resources`.

After adding/changing strings supplied to the goi18n `T()` translation func, run the following to update the translations binary:

    i18n4go -c fixup # answer any prompts appropriately
    goi18n -outdir cf/i18n/resources cf/i18n/resources/*.all.json
    bin/generate-language-resources

When running `i18n4go -c fixup`, you will be presented with the choices `new` or `upd` for each addition or update. Type in the appropriate choice. If `upd` is chosen, you will be asked to confirm which string is being updated using a numbered list.

After running the above, be sure to commit the translations binary, `cf/resources/i18n_resources.go`.

## Current Conventions

### Creating Commands

Resources that include several commands have been broken out into their own sub-package using the Resource name. An example of this convention is the Space resource and package (see `cf/commands/space`)

In addition, command file and methods naming follows a CRUD like convention. For example, the Space resource includes commands such a CreateSpace, ListSpaces, DeleteSpace, etc.

### Creating Repositories

Although not ideal, we use the name "Repository" for API related operations as opposed to "Service". Repository was chosen
to avoid confusion with Service model objects (i.e. creating Services and Service Instances within Cloud Foundry).

By convention, Repository methods return a model object and an error. Models are used in both Commands and Repositories
to model Cloud Foundry data. This convention provides a consistent method signature across repositories.
