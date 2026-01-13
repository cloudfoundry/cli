

<img src="https://raw.githubusercontent.com/cloudfoundry/logos/master/CF_Icon_4-colour.png" alt="CF logo" height="100" align="left"/>

# Cloud Foundry CLI
The official command line client for [Cloud Foundry](https://cloudfoundry.org).

View the latest help for [**The v8 CLI**](https://cli.cloudfoundry.org/en-US/v8) -OR- [**The v7 CLI**](https://cli.cloudfoundry.org/en-US/v7), or run `cf help -a` to view the help for all commands available in your currently installed version.

[![GitHub version](https://badge.fury.io/gh/cloudfoundry%2Fcli.svg)](https://github.com/cloudfoundry/cli/releases/latest)
[![Documentation](https://img.shields.io/badge/docs-online-ff69b4.svg)](https://docs.cloudfoundry.org/cf-cli)
[![Command help pages](https://img.shields.io/badge/command-help-lightgrey.svg)](https://cli.cloudfoundry.org)
[![Slack](https://slack.cloudfoundry.org/badge.svg)](https://slack.cloudfoundry.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/cloudfoundry/cli/blob/main/LICENSE)

CF CLI Binary Download Server's uptime:

[![Downloads Uptime](https://uptime.com/devices/services/widget/689896/c6d4bb7ddd16186d/service?light)](https://uptime.com/devices/services/689896/01026e1a663caab4)

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
Download and install the cf CLI from the [Downloads Section](#downloads) for the [v8 cf CLI](https://github.com/cloudfoundry/cli/wiki/V8-CLI-Installation-Guide).

Once installed, you can log in and push an app.
The currently supported version of the cf CLI:
1. The v8 cf CLI is backed by the [v3 CC API](http://v3-apidocs.cloudfoundry.org/version/3.85.0/) (with the exception of plugins). See [our v8 documentation](https://docs.cloudfoundry.org/cf-cli/v8.html) for more information.

**Note**: v7 and v6 CLI versions are no longer maintained or released.

View our [CLI v7 & v8 Versioning and Support Policy](https://github.com/cloudfoundry/cli/wiki/Versioning-and-Support-Policy) documentation.

If you have any questions, ask away on the #cli channel in [our Slack
community](https://slack.cloudfoundry.org/) and the
[cf-dev](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/)
mailing list, or [open a GitHub issue](https://github.com/cloudfoundry/cli/issues/new).  

## Contributing & Build Instructions
Please read the [contributors' guide](.github/CONTRIBUTING.md)

If you'd like to submit updated translations, please see the [i18n README](https://github.com/cloudfoundry/cli/blob/main/cf/i18n/README-i18n.md) for instructions on how to submit an update.

![Example](.github/cf_example.gif)

Check out our [community contributed CLI plugins](https://plugins.cloudfoundry.org) to further enhance your CLI experience.

## Downloads

### Installation instructions
- [Install V8](https://github.com/cloudfoundry/cli/wiki/V8-CLI-Installation-Guide)
- [Install V7](https://github.com/cloudfoundry/cli/wiki/V7-CLI-Installation-Guide) (**DEPRECATED** - no longer maintained)
- [Switching Between Multiple Versions](https://github.com/cloudfoundry/cli/wiki/Version-Switching-Guide)

## Known Issues
**Note:** For most up-to-date information in issues and workarounds please review [the open and closed github issues](https://github.com/cloudfoundry/cli/issues)

* On Windows in Cygwin and Git Bash, interactive password prompts (in `cf login`) do not hide the password properly from stdout ([issue #1835](https://github.com/cloudfoundry/cli/issues/1835)). Please use an alternative command (non-interactive authentication `cf auth` instead of `cf login`) to work around this. Or, use the Windows `cmd` command line.
* On Windows, `cf ssh` may not display correctly if the `TERM` is not set. We've found that setting `TERM` to `msys` fixes some of these issues.
* On Windows, `cf ssh` will hang when run from the MINGW32 or MINGW64 shell. A workaround is to use PowerShell instead.
* CF CLI/GoLang do not use OpenSSL. Custom/Self Signed Certificates need to be [installed in specific locations](https://docs.cloudfoundry.org/cf-cli/self-signed.html) in order to `login`/`auth` without `--skip-ssl-validation`.
* API tracing to terminal (using `CF_TRACE=true`, `-v` option or `cf config --trace`) doesn't work well with some CLI plugin commands. Trace to file works fine. On Linux, `CF_TRACE=/dev/stdout` works too. See [this Diego-Enabler plugin issue](https://github.com/cloudfoundry-attic/Diego-Enabler/issues/6) for more information.
* .cfignore used in `cf push` must be in UTF-8 encoding for CLI to interpret correctly. ([issue #281](https://github.com/cloudfoundry/cli/issues/281#issuecomment-65315518))
* On Linux, when encountering message "bash: .cf: No such file or directory", ensure that you're using the [correct binary or installer for your architecture](https://askubuntu.com/questions/133389/no-such-file-or-directory-but-the-file-exists).
* X-Cf-Warnings are printed through the `stdout`, if that's an inconvenience you could set `CF_RAISE_ERROR_ON_WARNINGS` and in that case warnings will be printed through the `stderr`. See [X-Cf-Warnings printed through stdout issue](https://github.com/cloudfoundry/cli/issues/2164)
* False negative message for user org creation. CLI v7.0 and CLI v7.1 non-admin users with the user-org-creation feature flag enabled will experience a failure when running cf create-org. The command will explicitly fail attempting to grant the user an org-manager role. However, it actually succeeds because the user would have an org-manager role granted to them via CAPI and therefore be able to access their org. This issue is resolved as of CLI v7.2. See [Inconsistent v2/v3 behavior around creating new orgs + assigning roles](https://github.com/cloudfoundry/cloud_controller_ng/issues/1879). 

## Filing Issues & Feature Requests

First, update to the [latest cli](https://github.com/cloudfoundry/cli/releases)
and try the command again.

If the error remains or feature still missing, check the [open issues](https://github.com/cloudfoundry/cli/issues) and if not already raised please file a new issue with the requested details.

## Plugin Development

The CF CLI supports external code execution via the plugins API. For more
information follow:

* [The CF CLI plugin development guide](https://github.com/cloudfoundry/cli/tree/main/plugin/plugin_examples)
* [The official plugins repository](https://plugins.cloudfoundry.org/)

When importing the plugin code use `import "code.cloudfoundry.org/cli/v9/plugin"`.
Older plugins that import `github.com/cloudfoundry/cli/plugin` will still work
as long they vendor the plugins directory.
