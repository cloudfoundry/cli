Cloud Foundry CLI [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli)
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

**Experimental:** Install CF for OSX through [Homebrew](http://brew.sh/) via the [pivotal's homebrew-tap](https://github.com/pivotal/homebrew-tap):

```
$ brew tap pivotal/tap
$ brew install cloudfoundry-cli
```

Also, edge binaries are published for [Mac OS X 64 bit](https://cli.run.pivotal.io/edge?arch=macosx64&source=github), [Windows 64 bit](https://cli.run.pivotal.io/edge?arch=windows64&source=github) and [Linux 64 bit](https://cli.run.pivotal.io/edge?arch=linux64&source=github) with each new 'push' that passes though CI.
These binaries are *not intended for wider use*; they're for developers to test new features and fixes as they are completed.

**Releases:** 32 bit releases and information about all our releases can be found [here](https://github.com/cloudfoundry/cli/releases)

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

Forking the repository for development
======================================

1. Install [Go](https://golang.org)
  * on Mac OS X this could be `brew install go` if you have [Homebrew](http://brew.sh/) installed.
1. Create a directory for your source & dependencies ex. `mkdir ~/dev/cloudfoundry-dev && cd ~/dev/cloudfoundry-dev`
1. [Ensure your $GOPATH is set correctly](http://golang.org/cmd/go/#hdr-GOPATH_environment_variable).  ex. `export GOPATH=~/dev/cloudfoundry-dev`
1. Install [godep](https://github.com/tools/godep) (usually `go get github.com/tools/godep`)
1. Get the cli source code: `go get github.com/cloudfoundry/cli`
  * (Ignore any warnings about "no buildable Go source files")
1. Run `godep restore` (note: this will modify the dependencies in your $GOPATH)
1. Fork the repository on GitHub
1. Add your fork as a remote: `cd $GOPATH/src/github.com/cloudfoundry/cli && git remote add your_name https://github.com/your_name/cli`
1. Optionally add build output location to your path: ex. `export PATH=$GOPATH/src/github.com/cloudfoundry/cli/out/:$PATH`

Building
========
To prepare your build environment, run `go get -u github.com/jteeuwen/go-bindata/...`

1. Run `./bin/build`
1. The binary will be built into the `./out` directory.

Optionally, you can use `bin/run` to compile and run the executable in one step.

If you want to run the tests with `ginkgo`, or build with `go build` you should first run `bin/generate-language-resources`. `bin/build` and `bin/test` generate language files automatically.

Developing
==========

1. Install [Mercurial](http://mercurial.selenic.com/)
1. Run `go get golang.org/x/tools/cmd/vet`
1. Write a Ginkgo test.
1. Run `bin/test` and watch the test fail.
1. Make the test pass.
1. Submit a pull request to the `master` branch.

**_*_ For development guide on writing a cli plugin, see [here](https://github.com/cloudfoundry/cli/tree/master/plugin_examples)**


Contributing
============

Major new feature proposals are given as a publically viewable google document with commenting allowed and discussed on the [cf-dev](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/) mailing list.

Pull Requests
---------------------

Pull Requests should be made against the `master` branch.

Architecture overview
---------------------

A command is a struct that implements this interface:

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
