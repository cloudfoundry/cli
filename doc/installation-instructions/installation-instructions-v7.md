### Downloading the latest V7 CF CLI

**DEPRECATED**: The v7 CF CLI is no longer maintained or released. Please consider upgrading to v8 or later versions. This documentation is kept for historical reference only.

**Important Note**: The v7 CF CLI is now GA and the binary has been renamed from `cf7` to `cf`. If you're already using the v7 CLI in parallel to the v6 CLI, you may need to change your workflow to accomodate the binary name change. See the [Version Switching](#version-switching) section for instructions. For more information and general status on the v7 CLI, please check [releases](https://github.com/cloudfoundry/cli/releases).

#### Compatibility
The v7 CLI's minimum supported version of the CC API is `v3.85.0` (published in [CAPI release v1.95.0](https://github.com/cloudfoundry/capi-release/releases/tag/1.95.0)).
See the [releases](https://github.com/cloudfoundry/cli/releases) page for the minimum CAPI versions required for each V7 release.


#### Installing using a package manager

**Mac OS X** and **Linux** using [Homebrew](https://brew.sh/) via the [cloudfoundry tap](https://github.com/cloudfoundry/homebrew-tap):

```sh
brew install cloudfoundry/tap/cf-cli@7
```

**Note:** `cf` tab completion requires `bash-completion` to be installed properly in order to work.

**Debian** and **Ubuntu** based Linux distributions:

```sh
# ...first add the Cloud Foundry Foundation public key and package repository to your system
wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key | sudo apt-key add -
echo "deb https://packages.cloudfoundry.org/debian stable main" | sudo tee /etc/apt/sources.list.d/cloudfoundry-cli.list
# ...then, update your local package index, then finally install the cf CLI
sudo apt-get update
sudo apt-get install cf7-cli
```

**Enterprise Linux** and **Fedora** systems (RHEL6/CentOS6 and up):
```sh
# ...first configure the Cloud Foundry Foundation package repository
sudo wget -O /etc/yum.repos.d/cloudfoundry-cli.repo https://packages.cloudfoundry.org/fedora/cloudfoundry-cli.repo
# ...then, install the cf CLI (which will also download and add the public key to your system)
sudo yum install cf7-cli
```


#### Installers and compressed binaries

| | Mac OS X 64 bit | Windows 64 bit | Linux 64 bit |
| :---------------: | :---------------: |:---------------:| :------------:|
| Installers |[pkg](https://packages.cloudfoundry.org/stable?release=macosx64&version=v7&source=github) | [zip](https://packages.cloudfoundry.org/stable?release=windows64&version=v7&source=github) | [rpm](https://packages.cloudfoundry.org/stable?release=redhat64&version=v7&source=github) / [deb](https://packages.cloudfoundry.org/stable?release=debian64&version=v7&source=github) |
| Binaries | [tgz](https://packages.cloudfoundry.org/stable?release=macosx64-binary&version=v7&source=github) |[zip](https://packages.cloudfoundry.org/stable?release=windows64-exe&version=v7&source=github)  | [tgz](https://packages.cloudfoundry.org/stable?release=linux64-binary&version=v7&source=github) |

Release notes, and 32 bit releases can be found [here](https://github.com/cloudfoundry/cli/releases).

**Download examples** with curl for Mac OS X and Linux binaries
```sh
# ...download & extract Mac OS X binary
curl -L "https://packages.cloudfoundry.org/stable?release=macosx64-binary&version=v7&source=github" | tar -zx
# ...or Linux 64-bit binary
curl -L "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=v7&source=github" | tar -zx
# ...move it to /usr/local/bin or a location you know is in your $PATH
mv cf /usr/local/bin
# ...copy tab completion file on Ubuntu (takes affect after re-opening your shell)
sudo curl -o /usr/share/bash-completion/completions/cf7 https://raw.githubusercontent.com/cloudfoundry/cli-ci/main/ci/installers/completion/cf7
# ...and to confirm your cf CLI version
cf version
```

##### Edge binaries
Edge binaries are *not intended for wider use*; they're for developers to test new features and fixes as they are 'pushed' and passed through the CI.
Follow these download links for [Mac OS X 64 bit](https://packages.cloudfoundry.org/edge?arch=macosx64&version=v7&source=github), [Windows 64 bit](https://packages.cloudfoundry.org/edge?arch=windows64&version=v7&source=github) and [Linux 64 bit](https://packages.cloudfoundry.org/edge?arch=linux64&version=v7&source=github).
