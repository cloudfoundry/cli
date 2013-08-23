Cloud Foundry CLI written in Go
===========

Goals
===========

Spike on converting the Cloud Foundry CLI tool to Go.

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
