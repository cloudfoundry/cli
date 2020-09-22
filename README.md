

<img src="https://raw.githubusercontent.com/cloudfoundry/logos/master/CF_Icon_4-colour.png" alt="CF logo" height="100" align="left"/>

# Cloud Foundry CLI
The official command line client for [Cloud Foundry](https://cloudfoundry.org).

View the latest help for [**The v7 CLI**](https://cli.cloudfoundry.org/en-US/v7) -OR- [**The v6 CLI**](https://cli.cloudfoundry.org/en-US/v6), or run `cf help -a` to view the help for all commands available in your currently installed version.

[![GitHub version](https://badge.fury.io/gh/cloudfoundry%2Fcli.svg)](https://github.com/cloudfoundry/cli/releases/latest)
[![Documentation](https://img.shields.io/badge/docs-online-ff69b4.svg)](https://docs.cloudfoundry.org/cf-cli)
[![Command help pages](https://img.shields.io/badge/command-help-lightgrey.svg)](https://cli.cloudfoundry.org)
[![Slack](https://slack.cloudfoundry.org/badge.svg)](https://slack.cloudfoundry.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/cloudfoundry/cli/blob/master/LICENSE)

***
<p align="left">
<b>Sections: </b>
<b><a href="#getting-started">Getting Started</a></b>
|
<b><a href="#downloads">Download</a></b>
|
<b><a href="#known-issues">Known Issues</a></b>
|
<b><a href="#filing-issues--feature-requests">Bugs/Feature Requests</a></b>
|
<b><a href="#plugin-development">Plugin Development</a></b>
|
<b><a href="#contributing--build-instructions">Contributing</a></b>
</p>

***
## Getting Started
Download and install the cf CLI from the [Downloads Section](#downloads) for either the [v7 cf CLI](https://github.com/cloudfoundry/cli/wiki/V7-CLI-Installation-Guide) or the [v6 cf CLI](https://github.com/cloudfoundry/cli/wiki/V6-CLI-Installation-Guide).

Once installed, you can log in and push an app.
There are now two supported versions of the cf CLI:
1. The v7 cf CLI is backed by the [v3 CC API](http://v3-apidocs.cloudfoundry.org/version/3.85.0/) (with the exception of plugins which will be migrated in the next major release). See [our v7 documentation](https://docs.cloudfoundry.org/cf-cli/v7.html) for more information.
1. The v6 cf CLI is backed by the [v2 CC API](https://apidocs.cloudfoundry.org/13.5.0/) See [our v6 documentation](https://docs.cloudfoundry.org/cf-cli) for more information.

## Contributing & Build Instructions
Please read the [contributors' guide](.github/CONTRIBUTING.md)

If you'd like to submit updated translations, please see the [i18n README](https://github.com/cloudfoundry/cli/blob/master/cf/i18n/README-i18n.md) for instructions on how to submit an update.

## Announcement ==> The V7 CLI is now Generally Available!
### The initial GA of the v7 cf CLI is Opt-In  
- You can pull down the GA release of the v7 cf CLI and/or the latest v6 cf CLI via our supported package managers using the same processes that were in place prior to the v7 GA (no changes are required initially)
- If you've been pulling down the v7 CF CLI beta previously, you will notice the name of the v7 binary has changed from `cf7` to `cf`
- See our [Version Switching Guide](https://github.com/cloudfoundry/cli/wiki/Version-Switching-Guide) for instructions on how to support workflows which require switching back and forth between v7 and v6

**A Note About Support**:</br>
Now that the v7 cf CLI is GA, all new features, enhancements, and fixes will be made on the v7 line.</br>
The v7 CLI's _minimum supported version_ of the CC API is `v3.85.0` (published in [CAPI release v1.95.0](https://github.com/cloudfoundry/capi-release/releases/tag/1.95.0)).</br>
The v7 CLI's _minimum supported version_ of the CF-Deployment is [`v13.5.0`](https://github.com/cloudfoundry/cf-deployment/releases/tag/v13.5.0).

Going forward, the v6 CLI will _only_ be updated to resolve the most severe defects and/or security issues.
At some point in the future, the v2 CC API endpoint will be deprecated by CAPI (see the [v2 CC API deprecation plan](https://docs.google.com/document/d/1KFZogeeexOqFf13oKHloe2QAorLh9OqwQHp8JvBl9lY/edit?usp=sharing)) and the v6 CLI will be incompatible CAPI once a `capi-release` that deprecates the v2 endpoint has been published.
Until the v2 CC API is deprecated, you can expect the v6 CLI to be fully functional, however, the CLI team's CI/CD resources are now focused on the v7 CLI so the v6 CLIs official **_maximum supported version_** of the CC APIs are now capped at `v2.149.0` and `v3.84.0` (published in [CAPI release v1.94.0](https://github.com/cloudfoundry/capi-release/releases/tag/1.94.0)), and the V6 CLIs official **_maximum supported version_** of the CF-Deployment is now capped at [`v13.4.0`](https://github.com/cloudfoundry/cf-deployment/releases/tag/v13.4.0).
 
The v6 CLI's _minimum supported version_ of the CF-Deployment is [`v7.0.0`](https://github.com/cloudfoundry/cf-deployment/releases/tag/v7.0.0). If you are on an older version of CF Deployment, we recommend you upgrade to CF-Deployment v7.0.0+.

If you have any questions, ask away on the #cli channel in [our Slack
community](https://slack.cloudfoundry.org/) and the
[cf-dev](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/)
mailing list, or [open a GitHub issue](https://github.com/cloudfoundry/cli/issues/new).  You can follow our development progress
on [Core CF CLI Pivotal Tracker](https://www.pivotaltracker.com/n/projects/892938).

**Need to switch back and forth between CLI versions?**
See the [Version Switching](#version-switching) section for instructions.

![Example](.github/cf_example.gif)

Check out our [community contributed CLI plugins](https://plugins.cloudfoundry.org) to further enhance your CLI experience.

## Downloads

### Installation instructions
- [Install V6](https://github.com/cloudfoundry/cli/wiki/V6-CLI-Installation-Guide)
- [Install V7](https://github.com/cloudfoundry/cli/wiki/V7-CLI-Installation-Guide)
- [Switching Between Multiple Versions](https://github.com/cloudfoundry/cli/wiki/Version-Switching-Guide)

## Known Issues

* On Windows in Cygwin and Git Bash, interactive password prompts (in `cf login`) do not hide the password properly from stdout ([issue #1835](https://github.com/cloudfoundry/cli/issues/1835)). Please use an alternative command (non-interactive authentication `cf auth` instead of `cf login`) to work around this. Or, use the Windows `cmd` command line.
* On Windows, `cf ssh` may not display correctly if the `TERM` is not set. We've found that setting `TERM` to `msys` fixes some of these issues.
* On Windows, `cf ssh` will hang when run from the MINGW32 or MINGW64 shell. A workaround is to use PowerShell instead.
* CF CLI/GoLang do not use OpenSSL. Custom/Self Signed Certificates need to be [installed in specific locations](https://docs.cloudfoundry.org/cf-cli/self-signed.html) in order to `login`/`auth` without `--skip-ssl-validation`.
* API tracing to terminal (using `CF_TRACE=true`, `-v` option or `cf config --trace`) doesn't work well with some CLI plugin commands. Trace to file works fine. On Linux, `CF_TRACE=/dev/stdout` works too. See [this Diego-Enabler plugin issue](https://github.com/cloudfoundry-attic/Diego-Enabler/issues/6) for more information.
* .cfignore used in `cf push` must be in UTF-8 encoding for CLI to interpret correctly. ([issue #281](https://github.com/cloudfoundry/cli/issues/281#issuecomment-65315518))
* On Linux, when encountering message "bash: .cf: No such file or directory", ensure that you're using the [correct binary or installer for your architecture](https://askubuntu.com/questions/133389/no-such-file-or-directory-but-the-file-exists).

## Filing Issues & Feature Requests

First, update to the [latest cli](https://github.com/cloudfoundry/cli/releases)
and try the command again.

If the error remains or feature still missing, check the [open issues](https://github.com/cloudfoundry/cli/issues) and if not already raised please file a new issue with the requested details.

## Plugin Development

The CF CLI supports external code execution via the plugins API. For more
information follow:

* [The CF CLI plugin development guide](https://github.com/cloudfoundry/cli/tree/master/plugin/plugin_examples)
* [The official plugins repository](https://plugins.cloudfoundry.org/)

When importing the plugin code use `import "code.cloudfoundry.org/cli/plugin"`.
Older plugins that import `github.com/cloudfoundry/cli/plugin` will still work
as long they vendor the plugins directory.
