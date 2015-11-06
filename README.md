Cloud Foundry CLI [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli)
===

This is the official command line client for [Cloud Foundry](cloudfoundry.org).

You can follow our development progress on [Pivotal Tracker](https://www.pivotaltracker.com/s/projects/892938).

Installation
===

Before you [log in](https://docs.pivotal.io/pivotalcf/devguide/installcf/whats-new-v6.html#login) and [push](https://docs.pivotal.io/pivotalcf/devguide/installcf/whats-new-v6.html#push) an app, you'll need a release.

### Stable releases
Download a stable release from the [releases page](http://github.com/cloudfoundry/cli/releases). If you download a binary release, make sure it's in available in your `$PATH`. For more information on installation, see the [Installation Guide](https://docs.pivotal.io/pivotalcf/devguide/installcf/install-go-cli.html).

There is a [Homebrew](http://brew.sh) tap available, as well, that delivers stable releases:

```
$ brew tap pivotal/tap
$ brew install cloudfoundry-cli
```

### Edge releases
**Edge releases are where active development of `cf` occurs.** Only use an edge release if you're comfortable with potential instability!

Edge releases are available for [Mac OS X 10.7+](https://cli.run.pivotal.io/edge?arch=macosx64&source=github), [64-bit Windows](https://cli.run.pivotal.io/edge?arch=windows64&source=github) and [64-bit Linux](https://cli.run.pivotal.io/edge?arch=linux64&source=github).

### Installing from source

1. Install [Go](https://golang.org/dl)
1. Ensure your `$GOPATH` [is set correctly](http://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
1. Install [godep](go get github.com/tools/godep)
1. Get the cli source code: `go get github.com/cloudfoundry/cli` (ignore the "no buildable Go source files" warning)
1. Run `godep restore` in `$GOPATH/src/github.com/cloudfoundry/cli` (note: this will modify the dependencies in your $GOPATH)
1. Run `go get -u github.com/jteeuwen/go-bindata/...`
1. Run `./bin/build` in `$GOPATH/src/cloudfoundry/cli`
1. Copy the binary in `$GOPATH/src/cloudfoundry/cli/out` to a location in your `$PATH`, such as `/usr/local/bin/`

Optionally, you can use `bin/run` to compile and run the executable without installing it.

Getting Help
===

* You can find further documentation at the docs page for the CLI [here](http://docs.cloudfoundry.org/devguide/#cf).
* There is also help available in the CLI itself; type `cf help` for more information.
* Each command also has help output available via `cf [command] --help` or `cf [command] -h`.
* For development guide on writing a cli plugin, see [here](https://github.com/cloudfoundry/cli/tree/master/plugin_examples).
* Finally, if you are still stuck or have any questions or issues, feel free to open a GitHub issue.

Troubleshooting / FAQs
===

Known Issues
------------
* .cfignore used in `cf push` must be in UTF8 encoding for CLI to interpret correctly.

Linux
-----
* "bash: .cf: No such file or directory". Ensure that you're using the correct binary or installer for your architecture. See http://askubuntu.com/questions/133389/no-such-file-or-directory-but-the-file-exists

Creating Issues
===========

##### For simple issues (eg: text formatting, help messages, etc), please provide

- the command you ran
- what occurred
- what you expected to occur

##### For issues related to HTTP requests or strange behavior, please run the command with env var `CF_TRACE=true` and provide


- the command you ran
- the trace output (**Make sure to REDACT any secrets in the log!**)
- a high-level description of the issue


##### For panics and other crashes, please provide

- the command you ran
- the stack trace generated (if any, making sure to REDACT any secrets in the log)
- any other relevant information

Contributing
===

Major new feature proposals are given as a publically viewable google document with commenting allowed and discussed on the [cf-dev](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/) mailing list.

1. Install [Go](https://golang.org)
1. [Ensure your $GOPATH is set correctly](http://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
1. Get `godep`: `go get github.com/tools/godep`
1. Install [Mercurial](http://mercurial.selenic.com/) (for go vet)
1. Get `go vet`: `go get golang.org/x/tools/cmd/vet`
1. Get the `cf` source: `go get github.com/cloudfoundry/cli`
  * (Ignore any warnings about "no buildable Go source files")
1. Run `godep restore` in $GOPATH/src/cloudfoundry/cli (note: this will modify the dependencies in your $GOPATH)
1. [Fork this repository](https://help.github.com/articles/fork-a-repo/), adding your fork as a remote
1. Write a new test, see it fail when running `bin/test` (or `ginkgo -p path/to/the/package/being/tested`)
1. Write code to pass the test
1. Repeat the above two steps until the feature is complete
1. Submit a [pull request](https://help.github.com/articles/using-pull-requests/) to the `master` branch

**_*_ For development guide on writing a cli plugin, see [here](https://github.com/cloudfoundry/cli/tree/master/plugin_examples)**

Architecture overview
===
A command is a struct that implements the Command interface:

```Go
type Command interface {
	MetaData() CommandMetadata
	SetDependency(deps Dependency, pluginCall bool) Command
	Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) (reqs []requirements.Requirement, err error)
	Execute(context flags.FlagContext)
}
```
[Source code](https://github.com/cloudfoundry/cli/blob/master/cf/command_registry/command.go#L9)

`Metadata()` is just a description of the command name, usage and flags:
```Go
type CommandMetadata struct {
	Name            string
	ShortName       string
	Usage           string
	Description     string
	Flags           map[string]flags.FlagSet
	SkipFlagParsing bool
	TotalArgs       int
}
```
[Source code](https://github.com/cloudfoundry/cli/blob/master/cf/command_registry/command.go#L16)

`Requirements()` returns a list of requirements that need to be met before a command can be invoked.

`Execute()` is the method that your command implements to do whatever it's supposed to do. The `context` object
provides flags and arguments.

When the command is run, it communicates with api using repositories (they are in [`cf/api`](https://github.com/cloudfoundry/cli/blob/master/cf/api)).

`SetDependency()` is where a command obtains its dependencies. Dependencies are typically declared as an interface type, and not a concrete type, so tests can inject a fake.
The bool argument `pluginCall` indicates whether the command is invoked by one of the CLI's plugin API methods.

Dependencies are injected into each command, so tests can inject a fake. This means that dependencies are
typically declared as an interface type, and not a concrete type. (see [`cf/command_registry/dependency.go`](https://github.com/cloudfoundry/cli/blob/master/cf/command_registry/dependency.go))

Some dependencies are managed by a repository locator in [`cf/api/repository_locator.go`](https://github.com/cloudfoundry/cli/blob/master/cf/api/repository_locator.go).

Repositories communicate with the api endpoints through a Gateway (see [`cf/net`](https://github.com/cloudfoundry/cli/tree/master/cf/net)).

Models are data structures related to Cloud Foundry (see [`cf/models`](https://github.com/cloudfoundry/cli/tree/master/cf/models)). For example, some models are
apps, buildpacks, domains, etc.


Managing dependencies
---------------------

Command dependencies are managed by the command registry package. The app uses the package (in [`cf/command_registry/dependency.go`](https://github.com/cloudfoundry/cli/blob/master/cf/command_registry/dependency.go))to instantiate them, this allows not sharing the knowledge of their dependencies with the app itself.

For commands that use another command as dependency, `command_registry` is used for retrieving the command dependency. For example, the command `restart` has a dependency on command `start` and `stop`, and this is how the command dependency is retrieved: [`restart.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commands/application/restart.go#L59)

As for repositories, we use the repository locator to handle their dependencies. You can find it in [`cf/api/repository_locator.go`](https://github.com/cloudfoundry/cli/blob/master/cf/api/repository_locator.go).

Example command
---------------

Create Space is a good example of a command. Its tests include checking arguments, requiring the user
to be logged in, and the actual behavior of the command itself. You can find it in [`cf/commands/space/create_space.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commands/space/create_space.go).

i18n
----
All pull requests which include user-facing strings should include updated translation files. These files are generated/ maintained using [i18n4go](https://github.com/maximilien/i18n4go). 

To add/ update translation strings run the command `i18n4go -c fixup`. For each change or update, you will be presented with the choices `new` or `upd`. Type in the appropriate choice. If `upd` is chosen, you will be asked to confirm which string is being updated using a numbered list.


Current conventions
===================

Creating Commands
-----------------

Resources that include several commands have been broken out into their own sub-package using the Resource name. An example of this convention is the Space resource and package (see `cf/commands/space`)

In addition, command file and methods naming follows a CRUD like convention. For example, the Space resource includes commands such a CreateSpace, ListSpaces, DeleteSpace, etc.

Creating Repositories
---------------------

Although not ideal, we use the name "Repository" for API related operations as opposed to "Service". Repository was chosen
to avoid confusion with Service model objects (i.e. creating Services and Service Instances within Cloud Foundry).

By convention, Repository methods return a model object and an error. Models are used in both Commands and Repositories
to model Cloud Foundry data. This convention provides a consistent method signature across repositories.
