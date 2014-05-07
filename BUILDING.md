Building Cloud Foundry CLI
==========================

For developing on unix systems:

1. Run `./bin/build`
1. The binary will be built into the `./out` directory.

Optionally, you can use `bin/run` to compile and run the executable in one step.

For developing on windows with powershell.exe:
1. $Env:GODEP_PATH=C:\path\to\go-path\src\github.com\cloudfoundry\cli\Godeps\_workspace;
1. $Env:GOPATH = $Env:GODEP_PATH + ";" + "C:\path\to\go-path\"

Building Installers and Cross Compiling On Unix Systems
=======================================================
1. [Configure your go installation for cross compilation](https://stackoverflow.com/questions/12168873/cross-compile-go-on-osx)
1. Run `bin/build-all.sh`
1. Run `ci/scripts/build-installers`
1. Installers will all be in the `release` dir

How We Test, Build, and Release The CLI
=======================================

High Level Overview
-------------------
Every push to the master branch goes through a CI pipeline that consists of

* unit tests
* integration tests

We run all of our tests on multiple platforms (e.g.: Linux, OS X, Windows) and
on multiple architectures (eg: 32bit, 64bit). Edge builds and tagged releases
are only released when all tests pass.

Unit Tests
----------
The first stage of every build is to run `bin/test` on all unix platforms (e.g.: 64 and 32bit Linux and OS X) and to
run an equivalent `go test` command on Windows. The executables produced by `go build` from this stage are uploaded
so that they can be run through integration tests and ultimately packaged into installers. This ensures that the
final products are fully tested and known to have passed our entire CI process.

The `ci/scripts` directory contains scripts that run tests and save the executable for each platform-architecture combination.

CATS
----
The [cf-acceptance-tests](https://github.com/cloudfoundry/cf-acceptance-tests) (eg: C.A.T.S.) are a suite of integration tests that
drive the `cf` cli along with a real CF deployment to verify the entire system works. We have some moderate tooling
to run these on different platforms, refer to the `herd-cats-$PLATFORM-$ARCH` scripts in `ci/scripts` for more
information.


GATS
----
The CLI team identified a need for integration tests *similar* to the CATS that we maintain; we call these tests
[GATS](https://github.com/tjarratt/GATS) (e.g.: GCF Acceptance Test Suite). These are run after the CATS tests,
and are fairly simple to run:

```
cd path/to/GATS

export API=http://api.some.ip.v4.address.xip.io
export ADMIN_USER=admin-user
export ADMIN_PASSWORD=admin-password
export CF_USER=user-name
export CF_USER_PASSWORD=user-password
export ORG=org-name
export SPACE=space-name
export APP_HOST=persistent-app-host

bin/configure
bin/test
```


Build and Release to S3
-----------------------
At the very end of our pipeline, assuming all tests have passed, we run a fairly simple script that uploads our
binaries and installers to the appropriate bucket on S3.

```
export AWS_SECRET_ACCESS_KEY=SECRET_KEY_IS_SECRET
export AWS_ACCESS_KEY_ID=WINK

ci/scripts/build-and-release
```

This script fetches the binaries that were produced earlier, generates installers for our supported platforms
and then uploads the final artifacts to S3.

Tagged Releases On Github
-------------------------
Every time we push to the master branch, a release is created in a directory in the go-cli bucket on our S3 account.
We make these URLs public so that people can try the edge builds.  Refer to our README for the URLs for some of these artifacts.

Commits that have a release tag on them (e.g.: v6.1.0) go into special directories that have the release name in them.

e.g.: http://go-cli.s3-website-us-east-1.amazonaws.com/releases/v6.1.0/
