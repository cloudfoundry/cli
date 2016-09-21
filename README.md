# Cloud Foundry CLI [![Apache License 2.0](https://img.shields.io/github/license/cloudfoundry/cli.svg)](https://github.com/cloudfoundry/cli/blob/master/LICENSE) [![Build Status](https://travis-ci.org/cloudfoundry/cli.png?branch=master)](https://travis-ci.org/cloudfoundry/cli) [![Code Climate](https://codeclimate.com/github/cloudfoundry/cli/badges/gpa.svg)](https://codeclimate.com/github/cloudfoundry/cli)

This is the official command line client for [Cloud Foundry](https://cloudfoundry.org).
Latest help of each command is [here](https://cli.cloudfoundry.org) (or run `cf help`);
Further documentation is at the [docs page for the
CLI](https://docs.cloudfoundry.org/cf-cli).  

If you have any questions, ask away on the #cli channel in [our Slack
community](http://slack.cloudfoundry.org/) and the
[cf-dev](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/)
mailing list, or open a GitHub issue.  You can follow our development progress
on [Pivotal Tracker](https://www.pivotaltracker.com/s/projects/892938).

## Getting Started

Download and run the installer for your platform from the [Downloads Section](#downloads).

Once installed, you can log in and push an app.
```
$ cf login -a api.[my-cloudfoundry].com
API endpoint: https://api.[my-cloudfoundry].com

Email> [my-email]

Password> [my-password]
Authenticating...
OK

$ cd [my-app-directory]
$ cf push
```

Check out our [community contributed CLI plugins](https://plugins.cloudfoundry.org) to further enhance your CLI experience.

## Downloads

**Latest stable:** Download the installer or compressed binary for your platform:

| | Mac OS X 64 bit | Windows 64 bit | Linux 64 bit |
| :---------------: | :---------------: |:---------------:| :------------:|
| Installers | [pkg](https://cli.run.pivotal.io/stable?release=macosx64&source=github) | [zip](https://cli.run.pivotal.io/stable?release=windows64&source=github) | [rpm](https://cli.run.pivotal.io/stable?release=redhat64&source=github) / [deb](https://cli.run.pivotal.io/stable?release=debian64&source=github) |
| Binaries | [tgz](https://cli.run.pivotal.io/stable?release=macosx64-binary&source=github) | [zip](https://cli.run.pivotal.io/stable?release=windows64-exe&source=github) | [tgz](https://cli.run.pivotal.io/stable?release=linux64-binary&source=github) |

**From the command line:** Download examples with curl for Mac OS X and Linux
```
# ...download & extract Mac OS X binary
$ curl -L "https://cli.run.pivotal.io/stable?release=macosx64-binary&source=github" | tar -zx
# ...or Linux 64-bit binary
$ curl -L "https://cli.run.pivotal.io/stable?release=linux64-binary&source=github" | tar -zx
# ...and to confirm your cf CLI version
$ ./cf --version
cf version x.y.z-...
```

**Via Homebrew:** Install CF for OSX through [Homebrew](http://brew.sh/) via the [cloudfoundry tap](https://github.com/cloudfoundry/homebrew-tap):

```
$ brew tap cloudfoundry/tap
$ brew install cf-cli
```

**Releases:** 32 bit releases and information about all our releases can be found [here](https://github.com/cloudfoundry/cli/releases).

Also, **edge binaries** are published for [Mac OS X 64 bit](https://cli.run.pivotal.io/edge?arch=macosx64&source=github), [Windows 64 bit](https://cli.run.pivotal.io/edge?arch=windows64&source=github) and [Linux 64 bit](https://cli.run.pivotal.io/edge?arch=linux64&source=github) with each new 'push' that passes though CI.
These binaries are *not intended for wider use*; they're for developers to test new features and fixes as they are completed.

## Known Issues

* In Cygwin and Git Bash on Windows, interactive prompts (such as in `cf login`) do not work (see #171). Please use alternative commands (e.g. `cf api` and `cf auth` to `cf login`) or option `-f` to suppress the prompts.
* .cfignore used in `cf push` must be in UTF8 encoding for CLI to interpret correctly.
* On Linux, when encountering message "bash: .cf: No such file or directory", ensure that you're using the correct binary or installer for your architecture. See http://askubuntu.com/questions/133389/no-such-file-or-directory-but-the-file-exists

## Filing Issues & Feature Requests

First, update to the [latest cli](https://github.com/cloudfoundry/cli/releases)
and try the command again.

If the error remains or feature still missing, check the [open issues](https://github.com/cloudfoundry/cli/issues) and if not already raised please file a new issue with the requested details.

## Plugin Development

For development guide on writing a cli plugin, see [here](https://github.com/cloudfoundry/cli/tree/master/plugin_examples).

## Contributing & Build Instructions

Please read the [contributors' guide](CONTRIBUTING.md)

If you'd like to submit updated translations, please see the [i18n README](https://github.com/cloudfoundry/cli/blob/master/cf/i18n/README-i18n.md) for instructions on how to submit an update.

