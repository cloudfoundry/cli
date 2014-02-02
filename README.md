Cloud Foundry CLI [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli)
=================

This is the official command line client for Cloud Foundry. [cf v6.0.0](https://github.com/cloudfoundry/cli/releases/tag/v6.0.0) is the current supported release. 

Downloading edge
================
Edge binaries are published to our Amazon S3 bucket with each new commit. These binaries are *not intended for wider use*, but for developers to test new features and fixes as they are completed:
- http://go-cli.s3.amazonaws.com/cf-darwin-amd64.tgz
- http://go-cli.s3.amazonaws.com/cf-linux-amd64.tgz
- http://go-cli.s3.amazonaws.com/cf-linux-386.tgz
- http://go-cli.s3.amazonaws.com/cf-windows-amd64.zip
- http://go-cli.s3.amazonaws.com/cf-windows-386.zip

You can follow our development progress on [Pivotal Tracker](https://www.pivotaltracker.com/s/projects/892938).

Cloning the repository
======================

1. Install [Go](http://golang.org)
1. Clone (Forking beforehand for development).
1. Run ```git submodule update --init --recursive```

Building
=======

1. Run ```./bin/build```
1. The binary will be built into the `./out` directory.

Developing
==========

1. Write a test.
1. Run ```bin/test``` and watch the test fail.
1. Make the test pass.
1. Submit a pull request.

If you want to run the benchmark tests

    ./bin/go test -bench . -benchmem cf/...

Contributing
============

Architecture overview
---------------------

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

Current conventions
===================

Creating commands
-----------------

Resources that include several commands have been broken out into their own sub-package using the Resource name. An example of this convention is the
Space resource and package.

In addition, command file and methods naming follows a CRUD like convention. For example, the Space resource includes commands such a CreateSpace, ListSpaces, etc.

Creating repositories
---------------------

Although not ideal, we use the name "Repository" for API related operations as opposed to "Service". Repository was chosen to avoid confusion with Service domain objects (i.e. creating Services and Service Instances within Cloud Foundry).

By convention, Repository methods return a Domain object and an ApiResponse. Domain objects are used in both Commands and Repositories to model Cloud Foundry data.  ApiResponse objects are used to communicate application errors, runtime errors, whether the resource was found, etc.
This convention provides a consistent method signature across repositories.
