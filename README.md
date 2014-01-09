Cloud Foundry CLI written in Go [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli)
===========

Background
===========

Project to rewrite the Cloud Foundry CLI tool using Go. This project should currently be considered alpha quality
software and should not be used in production environments. If you need something more stable, please check
out the [RubyGem](https://github.com/cloudfoundry/cf).

For a view on the current status of the project, check [cftracker](http://cftracker.cfapps.io/cfcli).

Cloning the repository
======================

1. Install Go ```brew install go --cross-compile-common```
1. Clone (Fork before hand for development).
1. Run ```git submodule update --init --recursive```

Downloading Edge
========
The latest binary builds and installers are published to Amazon S3 buckets.

Binaries:
- http://go-cli.s3.amazonaws.com/gcf-darwin-amd64.tgz
- http://go-cli.s3.amazonaws.com/gcf-linux-amd64.tgz
- http://go-cli.s3.amazonaws.com/gcf-linux-386.tgz
- http://go-cli.s3.amazonaws.com/gcf-windows-amd64.zip
- http://go-cli.s3.amazonaws.com/gcf-windows-386.zip

Installers:
- http://go-cli.s3.amazonaws.com/installer-windows-amd64.zip
- http://go-cli.s3.amazonaws.com/installer-windows-386.zip
- http://go-cli.s3.amazonaws.com/cf-cli_i386.deb
- http://go-cli.s3.amazonaws.com/cf-cli_i386.rpm

Building
========

1. Run ```./bin/build```
1. The binary will be built into the out directory.

Development
===========

NOTE: Currently only development on OSX 10.8 is supported

1. Write a test.
1. Run ``` bin/test ``` and watch test fail.
1. Make test pass.
1. Submit a pull request.

If you want to run the benchmark tests

    ./bin/go test -bench . -benchmem cf/...

Releasing
=========

run ```bin/build-all```

This will create tgz files and installers in the release folder.

Contributing
============

Rough overview of the architecture
----------------------------------

The app (in ```src/cf/app/app.go```) declares the list of available commands. Help and flags are defined there.
It will instantiate a command, and run it using the runner (in ```src/cf/commands/runner.go```).

A command has requirements, and a run function. Requirements are used as filters before running the command.
If any of them fails, the command will not run (see ```src/cf/requirements``` for examples of requirements).

When the command is run, it communicates with api using repositories (they are in ```src/cf/api```).

Repositories are injected into the command, so tests can inject a fake.

Repositories communicate with the api endpoints through a Gateway (see ```src/cf/net```).

Repositories return a Domain Object and an ApiResponse object.

Domain objects are data structures related to Cloud Foundry (see ```src/cf/domain```).

ApiResponse objects convey a variety of important error conditions (see ```src/cf/net/api_status```).


Managing dependencies
---------------------

Command dependencies are managed by the commands factory. The app uses the command factory (in ```src/cf/commands/factory.go```)
to instantiate them, this allows not sharing the knowledge of their dependencies with the app itself.

As for repositories, we use the repository locator to handle their dependencies. You can find it in ```src/cf/api/repository_locator.go```.

Example command
---------------

Create Space is a good example of command. Its tests include checking arguments, having requirements, and the actual command itself.
You will find it in ```src/cf/commands/space/create_space.go```.

Current Conventions
===================

Creating Commands
-----------------

Resources that include several commands have been broken out into their own sub-package using the Resource name. An example of this convention is the
Space resource and package.

In addition, command file and methods naming follows a CRUD like convention. For example, the Space resource includes commands such a CreateSpace, ListSpaces, etc.

Creating Repositories
---------------------

Although not ideal, we use the name "Repository" for API related operations as opposed to "Service". Repository was chosen to avoid confusion with Service domain objects (i.e. creating Services and Service Instances within Cloud Foundry).

By convention, Repository methods return a Domain object and an ApiResponse. Domain objects are used in both Commands and Repositories to model Cloud Foundry data.  ApiResponse objects are used to communicate application errors, runtime errors, whether the resource was found, etc.
This convention provides a consistent method signature across repositories.
