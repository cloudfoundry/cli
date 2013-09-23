Cloud Foundry CLI written in Go [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli)
===========

Background
===========

Project to rewrite the Cloud Foundry CLI tool using Go. This project should currently be considered alpha quality
software and should not be used in production environments. If you need something more stable, please check
out the [RubyGem](https://github.com/cloudfoundry/cf).

For a view on the current status of the project, check [cftracker](http://cftracker.cfapps.io/cfcli).

Building
========

1. Run ```./bin/build```
1. The binary will be built into the out directory.

Development
===========

NOTE: Currently only development on OSX 10.8 is supported

1. Install Go ```brew install go --cross-compile-common```
1. Fork and clone.
1. Run ```git submodule update --init --recursive```
1. Write a test.
1. Run ``` bin/test ``` and watch test fail.
1. Make test pass.
1. Submit a pull request.

Releasing
=========

1. Run ```bin/build-all```. This will create tgz files in the release folder.

Contributing
============

Rough overview of the architecture
----------------------------------

The app (in ```src/cf/app/app.go```) declares the list of available commands. Help and flags are defined there.
It will instantiate a command, and run it using the runner (in ```src/cf/commands/runner.go```).

A command has requirements, and a run function. Requirements are use as filters before running the command.
If any of them fails, the command will not run (see ```src/cf/requirements``` for examples of requirements).

When the command is run, it communicates with api using repositories (they are in ```src/cf/api```).
Repositories should be injected into the command, so that your tests can inject a fake.

Repositories communicate with the network layer, usually through a Gateway (see ```src/cf/net```).

Managing dependencies
---------------------

Commands dependencies are managed by the commands factory. The app uses the command factory (in ```src/cf/commands/factory.go```)
to instantiate them, this allows not sharing the knowledge of their dependencies with the app itself.

As for repositories, we use the repository locator to handle their dependencies. You can find it in ```src/cf/api/repository_locator.go```.

Example command
---------------

Create Space is a good example of command. Its tests include checking arguments, having requirements, and the actual command itself.
You will find it in ```src/cf/commands/create_space.go```.
