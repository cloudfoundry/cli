Cloud Foundry CLI written in Go [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli)
===========

Background
===========

Project to rewrite the Cloud Foundry CLI tool using Go. This project should currently be considered alpha quality
software and should not be used in production environments. If you need something more stable, please check
out the [RubyGem](https://github.com/cloudfoundry/cf).

Building
========
1. Run ```./bin/build```
1. The binary will be built into the `$GOPATH/bin` directory if `$GOPATH` is set; otherwise built into the `./out` directory.

Development
===========

NOTE: Currently only development on OSX 10.8 is supported

1. Install Go ```brew install go --cross-compile-common```
1. Setup `$GOLANG` env variable and add `$GOLANG/bin` to `$PATH`
1. Fork and clone: `git clone https://github.com/cloudfoundry/cli.git $GOPATH/src/github.com/cloudfoundry/cli`
1. Run ```git submodule update --init --recursive```
1. Write a test.
1. Run ``` bin/test ``` and watch test fail.
1. Make test pass.
1. Submit a pull request.

Releasing
=========

1. Run ```bin/build-all```. This will create tgz files in the release folder.
