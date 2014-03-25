Cloud Foundry CLI [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli)
=================

This is the official command line client for Cloud Foundry.

Stable Release
==============

Installers
----------
- [Debian 32 bit](https://s3.amazonaws.com/go-cli/releases/latest/cf-cli_i386.deb)
- [Debian 64 bit](https://s3.amazonaws.com/go-cli/releases/latest/cf-cli_amd64.deb)
- [Redhat 32 bit](https://s3.amazonaws.com/go-cli/releases/latest/cf-cli_i386.rpm)
- [Redhat 64 bit](https://s3.amazonaws.com/go-cli/releases/latest/cf-cli_amd64.rpm)
- [Mac OS X 64 bit](https://s3.amazonaws.com/go-cli/releases/latest/installer-osx-amd64.pkg)
- [Windows 32 bit](https://s3.amazonaws.com/go-cli/releases/latest/installer-windows-386.zip)
- [Windows 64 bit](https://s3.amazonaws.com/go-cli/releases/latest/installer-windows-amd64.zip)

Binaries
--------
- [Linux 32 bit binary](https://s3.amazonaws.com/go-cli/releases/latest/cf-linux-386.tgz)
- [Linux 64 bit binary](https://s3.amazonaws.com/go-cli/releases/latest/cf-linux-amd64.tgz)
- [Mac OS X 64 bit binary](https://s3.amazonaws.com/go-cli/releases/latest/cf-darwin-amd64.tgz)
- [Windows 32 bit binary](https://s3.amazonaws.com/go-cli/releases/latest/cf-windows-386.zip)
- [Windows 64 bit binary](https://s3.amazonaws.com/go-cli/releases/latest/cf-windows-amd64.zip)


Edge Releases (master)
======================

Edge binaries are published to our Amazon S3 bucket with each new commit that passes CI.
These binaries are *not intended for wider use*, but for developers to test new features and fixes as they are completed:

- [Linux 64 bit binary](http://go-cli.s3.amazonaws.com/master/cf-linux-amd64.tgz)
- [Linux 32 bit binary](http://go-cli.s3.amazonaws.com/master/cf-linux-386.tgz)
- [Mac OS X 64 bit binary](http://go-cli.s3.amazonaws.com/master/cf-darwin-amd64.tgz)
- [Windows 64 bit binary](http://go-cli.s3.amazonaws.com/master/cf-windows-amd64.zip)
- [Windows 32 bit binary](http://go-cli.s3.amazonaws.com/master/cf-windows-386.zip)

You can follow our development progress on [Pivotal Tracker](https://www.pivotaltracker.com/s/projects/892938).

Troubleshooting / FAQs
======================

Linux
-----
* "bash: .cf: No such file or directory". Ensure that you're using the correct binary or installer for your architecture. See http://askubuntu.com/questions/133389/no-such-file-or-directory-but-the-file-exists

Mac OS X
--------
* "Illegal instruction: 4" - This is a known defect with some versions of OS X. Please file an issue. As a workaround, you can build from source or use homebrew.


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

Cloning the repository
======================

1. Install [Go](http://golang.org)
1. Clone (Forking beforehand for development).
1. Run `git submodule update --init --recursive`

Building
=======

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

The app (in `src/cf/app/app.go`) declares the list of available commands, which are composed of a Name,
Description, Usage and any optional Flags. The action for each command is to instantiate a command object,
 which is invoked by the runner (in `src/cf/commands/runner.go`).

A command has `Requirements`, and a `Run` function. Requirements are used as filters before running the command.
If any of them fails, the command will not run (see `src/cf/requirements` for examples of requirements).

When the command is run, it communicates with api using repositories (they are in `src/cf/api`).

Dependencies are injected into each command, so tests can inject a fake. This means that dependencies are
typically declared as an interface type, and not a concrete type. (see `src/cf/commands/factory.go`)

Some dependencies are managed by a repository locator in `src/cf/api/repository_locator.go`.

Repositories communicate with the api endpoints through a Gateway (see `src/cf/net`).

Models are data structures related to Cloud Foundry (see `src/cf/models`). For example, some models are
apps, buildpacks, domains, etc.


Managing dependencies
---------------------

Command dependencies are managed by the commands factory. The app uses the command factory (in `src/cf/commands/factory.go`)
to instantiate them, this allows not sharing the knowledge of their dependencies with the app itself.

As for repositories, we use the repository locator to handle their dependencies. You can find it in `src/cf/api/repository_locator.go`.

Example command
---------------

Create Space is a good example of a command. Its tests include checking arguments, requiring the user
to be logged in, and the actual behavior of the command itself. You can find it in `src/cf/commands/space/create_space.go`.

Current conventions
===================

Creating Commands
-----------------

Resources that include several commands have been broken out into their own sub-package using the Resource name. An example
of this convention is the Space resource and package (see `src/cf/commands/space`)

In addition, command file and methods naming follows a CRUD like convention. For example, the Space resource includes commands
such a CreateSpace, ListSpaces, DeleteSpace, etc.

Creating Repositories
---------------------

Although not ideal, we use the name "Repository" for API related operations as opposed to "Service". Repository was chosen
to avoid confusion with Service model objects (i.e. creating Services and Service Instances within Cloud Foundry).

By convention, Repository methods return a model object and an error. Models are used in both Commands and Repositories
to model Cloud Foundry data. This convention provides a consistent method signature across repositories.
