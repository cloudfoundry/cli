Cloud Foundry CLI [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli)
=================

This is the official command line client for Cloud Foundry.

Getting Started
===============
Download and run the installer for your platform from the section below. 

If you are on OS X, you can also install the CLI with homebrew--run (Experimental)

```
$ brew tap pivotal/tap
$ brew install cloudfoundry-cli
```

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
You can find further documentation at the docs page for the CLI [here](http://docs.cloudfoundry.org/devguide/#cf).
There is also help available in the CLI itself; type `cf help` for more information.
Each command also has help output available via `cf [command] --help` or `cf [command] -h`.
Finally, if you are still stuck, feel free to open a GitHub issue.

Stable Release
==============

Installers
----------
- [Debian 32 bit](https://cli.run.pivotal.io/stable?release=debian32)
- [Debian 64 bit](https://cli.run.pivotal.io/stable?release=debian64)
- [Redhat 32 bit](https://cli.run.pivotal.io/stable?release=redhat32)
- [Redhat 64 bit](https://cli.run.pivotal.io/stable?release=redhat64)
- [Mac OS X 64 bit](https://cli.run.pivotal.io/stable?release=macosx64)
- [Windows 32 bit](https://cli.run.pivotal.io/stable?release=windows32)
- [Windows 64 bit](https://cli.run.pivotal.io/stable?release=windows64)

Edge Releases (master)
======================

Edge binaries are published to our Amazon S3 bucket with each new commit that passes CI.
These binaries are *not intended for wider use*; they're for developers to test new features and fixes as they are completed:

- [Linux 32 bit binary](https://cli.run.pivotal.io/edge?arch=linux32)
- [Linux 64 bit binary](https://cli.run.pivotal.io/edge?arch=linux64)
- [Mac OS X 64 bit binary](https://cli.run.pivotal.io/edge?arch=macosx64)
- [Windows 32 bit binary](https://cli.run.pivotal.io/edge?arch=windows32)
- [Windows 64 bit binary](https://cli.run.pivotal.io/edge?arch=windows64)

You can follow our development progress on [Pivotal Tracker](https://www.pivotaltracker.com/s/projects/892938).

Troubleshooting / FAQs
======================

Linux
-----
* "bash: .cf: No such file or directory". Ensure that you're using the correct binary or installer for your architecture. See http://askubuntu.com/questions/133389/no-such-file-or-directory-but-the-file-exists

Filing Bugs
===========

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
1. [Ensure your $GOPATH is set correctly](http://golang.org/cmd/go/#hdr-GOPATH_environment_variable)
1. Install [godep](https://github.com/tools/godep)
1. Get the cli source code: `go get github.com/cloudfoundry/cli`
  * (Ignore any warnings about "no buildable Go source files")
1. Fork the repository
1. Add your fork as a remote: `cd $GOPATH/src/github.com/cloudfoundry/cli && git remote add your_name https://github.com/your_name/cli`

Building
========

1. Run `./bin/build`
1. The binary will be built into the `./out` directory.

Optionally, you can use `bin/run` to compile and run the executable in one step.

Developing
==========

1. Install [Mercurial](http://mercurial.selenic.com/)
1. Run `go get code.google.com/p/go.tools/cmd/vet`
1. Write a Ginkgo test.
1. Run `bin/test` and watch the test fail.
1. Make the test pass.
1. Submit a pull request.

Contributing
============

Architecture overview
---------------------

A command is a struct that implements this interface:

```
type Command interface {
	Metadata() command_metadata.CommandMetadata
	GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error)
	Run(c *cli.Context)
}
```

`Metadata()` is just a description of the command name, usage and flags:
```
type CommandMetadata struct {
	Name            string
	ShortName       string
	Usage           string
	Description     string
	Flags           []cli.Flag
	SkipFlagParsing bool
}
```

`GetRequirements()` returns a list of requirements that need to be met before a command can be invoked.

`Run()` is the method that your command implements to do whatever it's supposed to do. The `context` object
provides flags and arguments.

When the command is run, it communicates with api using repositories (they are in `cf/api`).

Dependencies are injected into each command, so tests can inject a fake. This means that dependencies are
typically declared as an interface type, and not a concrete type. (see `cf/commands/factory.go`)

Some dependencies are managed by a repository locator in `cf/api/repository_locator.go`.

Repositories communicate with the api endpoints through a Gateway (see `cf/net`).

Models are data structures related to Cloud Foundry (see `cf/models`). For example, some models are
apps, buildpacks, domains, etc.


Managing dependencies
---------------------

Command dependencies are managed by the commands factory. The app uses the command factory (in `cf/commands/factory.go`)
to instantiate them, this allows not sharing the knowledge of their dependencies with the app itself.

As for repositories, we use the repository locator to handle their dependencies. You can find it in `cf/api/repository_locator.go`.

Example command
---------------

Create Space is a good example of a command. Its tests include checking arguments, requiring the user
to be logged in, and the actual behavior of the command itself. You can find it in `cf/commands/space/create_space.go`.

Current conventions
===================

Creating Commands
-----------------

Resources that include several commands have been broken out into their own sub-package using the Resource name. An example
of this convention is the Space resource and package (see `cf/commands/space`)

In addition, command file and methods naming follows a CRUD like convention. For example, the Space resource includes commands
such a CreateSpace, ListSpaces, DeleteSpace, etc.

Creating Repositories
---------------------

Although not ideal, we use the name "Repository" for API related operations as opposed to "Service". Repository was chosen
to avoid confusion with Service model objects (i.e. creating Services and Service Instances within Cloud Foundry).

By convention, Repository methods return a model object and an error. Models are used in both Commands and Repositories
to model Cloud Foundry data. This convention provides a consistent method signature across repositories.
