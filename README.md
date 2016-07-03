Cloud Foundry CLI [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli) [![Code Climate](https://codeclimate.com/github/cloudfoundry/cli/badges/gpa.svg)](https://codeclimate.com/github/cloudfoundry/cli)
=================

This is the official command line client for Cloud Foundry.

You can follow our development progress on [Pivotal Tracker](https://www.pivotaltracker.com/s/projects/892938).

Getting Started
===============
Download and run the installer for your platform from the [Downloads Section](#downloads).

Once installed, you can log in and push an app.
```
$ cd [my-app-directory]
$ cf api api.[my-cloudfoundry].com
Setting api endpoint to https://api.[my-cloudfoundry].com...
OK

$ cf login
API endpoint: https://api.[my-cloudfoundry].com

Email> [my-email]

Password> [my-password]
Authenticating...
OK

$ cf push
```
#Further Reading and Getting Help
* You can find further documentation at the docs page for the CLI [here](http://docs.cloudfoundry.org/devguide/#cf).
* There is also help available in the CLI itself; type `cf help` for more information.
* Each command also has help output available via `cf [command] --help` or `cf [command] -h`.   
* For development guide on writing a cli plugin, see [here](https://github.com/cloudfoundry/cli/tree/master/plugin_examples).  
* Finally, if you are still stuck or have any questions or issues, feel free to open a GitHub issue.

Downloads
=========
**Latest stable:** Download the installer or compressed binary for your platform:

| | Mac OS X 64 bit | Windows 64 bit | Linux 64 bit |
| :---------------: | :---------------: |:---------------:| :------------:|
| Installers | [pkg](https://cli.run.pivotal.io/stable?release=macosx64&source=github) | [zip](https://cli.run.pivotal.io/stable?release=windows64&source=github) | [rpm](https://cli.run.pivotal.io/stable?release=redhat64&source=github) / [deb](https://cli.run.pivotal.io/stable?release=debian64&source=github) |
| Binaries | [tgz](https://cli.run.pivotal.io/stable?release=macosx64-binary&source=github) | [zip](https://cli.run.pivotal.io/stable?release=windows64-exe&source=github) | [tgz](https://cli.run.pivotal.io/stable?release=linux64-binary&source=github) |

**From the command line:** Download examples with curl for Mac OS X and Linux
```
# ...download & extract Mac OS X binary
$ curl -L "https://cli.run.pivotal.io/stable?release=macosx64-binary&source=github" | tar -zx
# ...or Linux binary
$ curl -L "https://cli.run.pivotal.io/stable?release=linux64-binary&source=github" | tar -zx
# ...and confirm you got the version you expected
$ ./cf --version
cf version x.y.z-...
```

**Via Homebrew:** Install CF for OSX through [Homebrew](http://brew.sh/) via the [cloudfoundry tap](https://github.com/cloudfoundry/homebrew-tap):

```
$ brew tap cloudfoundry/tap
$ brew install cf-cli
```

### Edge Binaries
**We strongly discourage the use of Edge Binaries**. Edge binaries are for people who want to use new experimental features that may not work at all. Some functionality depends on the edge of CF, which most people do not have installed.

| Mac OS X 64 bit | Windows 64 bit | Linux 64 bit |
| :---------------: |:---------------:| :------------:|
| [tgz](https://cli.run.pivotal.io/edge?arch=macosx64&source=github) | [zip](https://cli.run.pivotal.io/edge?arch=windows64&source=github) | [tgz](https://cli.run.pivotal.io/edge?arch=linux64&source=github) |

### All Releases 
All our releases, including 32bit binaries, can be found [here](https://github.com/cloudfoundry/cli/releases)

Troubleshooting / FAQs
======================

Known Issues
------------
* .cfignore used in `cf push` must be in UTF8 encoding for CLI to interpret correctly.

Linux
-----
* "bash: .cf: No such file or directory". Ensure that you're using the correct binary or installer for your architecture. See http://askubuntu.com/questions/133389/no-such-file-or-directory-but-the-file-exists

Filing Bugs
===========

First, update to the [latest cli](https://github.com/cloudfoundry/cli/releases)
and try the command again.

If the error remains, run the command that exposes the bug with the environment
variable CF_TRACE set to true and [create an
issue](https://github.com/cloudfoundry/cli/issues).

Include the below information when creating the issue:

* The error that occurred
* The stack trace (if applicable)
* The command you ran (e.g. `cf org-users`)
* The CLI Version (e.g. 6.13.0-dfba612)
* Your platform details (e.g. Mac OS X 10.11, Windows 8.1 64-bit, Ubuntu 14.04.3 64-bit)
* The shell you used (e.g. Terminal, iTerm, Powershell, Cygwin, gnome-terminal, terminator)

##### For simple bugs (eg: text formatting, help messages, etc), please provide

- the command you ran
- what occurred
- what you expected to occur

##### For bugs related to HTTP requests or strange behavior, please run the command with env var `CF_TRACE=true` and provide

- the command you ran
- the trace output
- a high-level description of the bug

##### For panics and other crashes, please provide

- the command you ran
- the stack trace generated (if any)
- any other relevant information

Contributing
============

Major new feature proposals are given as a publically viewable google document with commenting allowed and discussed on the [cf-dev](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/) mailing list.

1. Install [Go 1.6.x](https://golang.org)
1. Create a directory where you would like to store the source for Go projects and their binaries (e.g. `$HOME/go`)
1. Set an environment variable, `GOPATH`, pointing at the directory you created
1. Get the `cf` source: `go get github.com/cloudfoundry/cli`
  * (Ignore any warnings about "no buildable Go source files")
1. [Fork this repository](https://help.github.com/articles/fork-a-repo/), adding your fork as a remote
1. Run our bootstrap script, `bin/bootstrap`
1. Write a new test, see it fail when running `bin/test` (or `ginkgo -p path/to/the/package/being/tested`)
1. Write code to pass the test
1. Repeat the above two steps until the feature is complete
1. Submit a [pull request](https://help.github.com/articles/using-pull-requests/) to the `master` branch

**_*_ For development guide on writing a cli plugin, see [here](https://github.com/cloudfoundry/cli/tree/master/plugin_examples)**

Architecture overview
---------------------

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


Managing dependencies
---------------------

Command dependencies are managed by the command registry package. The app uses the package (in [`cf/commandregistry/dependency.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commandregistry/dependency.go)) to instantiate them, this allows not sharing the knowledge of their dependencies with the app itself.

For commands that use another command as dependency, `commandregistry` is used for retrieving the command dependency. For example, the command `restart` has a dependency on command `start` and `stop`, and this is how the command dependency is retrieved: [`restart.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commands/application/restart.go#L59)

As for repositories, we use the repository locator to handle their dependencies. You can find it in [`cf/api/repository_locator.go`](https://github.com/cloudfoundry/cli/blob/master/cf/api/repository_locator.go).

Example command
---------------

Create Space is a good example of a command. Its tests include checking arguments, requiring the user
to be logged in, and the actual behavior of the command itself. You can find it in [`cf/commands/space/create_space.go`](https://github.com/cloudfoundry/cli/blob/master/cf/commands/space/create_space.go).

i18n
----
#### For Translators

If you'd like to submit updated translations, please see the [i18n README](https://github.com/cloudfoundry/cli/blob/master/cf/i18n/README-i18n.md) for instructions on how to submit an update.

#### For CLI Developers

If you are adding new strings or updating existing strings within the CLI code, you'll need to update the binary representation of the translation files. This file is generated/maintained using [i18n4go](https://github.com/krishicks/i18n4go), [goi18n](https://github.com/nicksnyder/go-i18n), and `bin/generate-language-resources`.

After adding/changing strings supplied to the goi18n `T()` translation func, run the following to update the translations binary:

    i18n4go -c fixup # answer any prompts appropriately
    goi18n -outdir cf/i18n/resources cf/i18n/resources/*.all.json
    bin/generate-language-resources

When running `i18n4go -c fixup`, you will be presented with the choices `new` or `upd` for each addition or update. Type in the appropriate choice. If `upd` is chosen, you will be asked to confirm which string is being updated using a numbered list.

After running the above, be sure to commit the translations binary, `cf/resources/i18n_resources.go`.

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
