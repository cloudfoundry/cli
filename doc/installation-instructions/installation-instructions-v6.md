### Downloading the latest V6 CF CLI

**DEPRECATED**: The v6 CF CLI is no longer maintained or released. Please consider upgrading to v8 or later versions. This documentation is kept for historical reference only.

#### Installing using a package manager

**Mac OS X** and **Linux** using [Homebrew](https://brew.sh/) via the [cloudfoundry tap](https://github.com/cloudfoundry/homebrew-tap):

```sh
brew install cloudfoundry/tap/cf-cli@6
```

**Note:** `cf` tab completion requires `bash-completion` to be installed properly in order to work.

**Debian** and **Ubuntu** based Linux distributions:

```sh
# ...first add the Cloud Foundry Foundation public key and package repository to your system
wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key | sudo apt-key add -
echo "deb https://packages.cloudfoundry.org/debian stable main" | sudo tee /etc/apt/sources.list.d/cloudfoundry-cli.list
# ...then, update your local package index, then finally install the cf CLI
sudo apt-get update
sudo apt-get install cf-cli
```

**Enterprise Linux** and **Fedora** systems (RHEL6/CentOS6 and up):
```sh
# ...first configure the Cloud Foundry Foundation package repository
sudo wget -O /etc/yum.repos.d/cloudfoundry-cli.repo https://packages.cloudfoundry.org/fedora/cloudfoundry-cli.repo
# ...then, install the cf CLI (which will also download and add the public key to your system)
sudo yum install cf-cli
```

#### Installers and compressed binaries


| | Mac OS X 64 bit | Windows 64 bit | Linux 64 bit |
| :---------------: | :---------------: |:---------------:| :------------:|
| Installers | [pkg](https://packages.cloudfoundry.org/stable?release=macosx64&source=github&version=v6) |[zip](https://packages.cloudfoundry.org/stable?release=windows64&source=github&version=v6)  | [rpm](https://packages.cloudfoundry.org/stable?release=redhat64&source=github&version=v6) / [deb](https://packages.cloudfoundry.org/stable?release=debian64&source=github&version=v6) |
| Binaries | [tgz](https://packages.cloudfoundry.org/stable?release=macosx64-binary&source=github&version=v6) | [zip](https://packages.cloudfoundry.org/stable?release=windows64-exe&source=github&version=v6) | [tgz](https://packages.cloudfoundry.org/stable?release=linux64-binary&source=github&version=v6) |

Release notes, and 32 bit releases can be found [here](https://github.com/cloudfoundry/cli/releases).

**Download examples** with curl for Mac OS X and Linux binaries
```sh
# ...download & extract Mac OS X binary
curl -L "https://packages.cloudfoundry.org/stable?release=macosx64-binary&source=github&version=v6" | tar -zx
# ...or Linux 64-bit binary
curl -L "https://packages.cloudfoundry.org/stable?release=linux64-binary&source=github&version=v6" | tar -zx
# ...move it to /usr/local/bin or a location you know is in your $PATH
mv cf /usr/local/bin
# ...copy tab completion file on Ubuntu (takes affect after re-opening your shell)
sudo curl -o /usr/share/bash-completion/completions/cf https://raw.githubusercontent.com/cloudfoundry/cli-ci/main/ci/installers/completion/cf
# ...and to confirm your cf CLI version
cf version
```

##### Edge binaries
Edge binaries are *not intended for wider use*; they're for developers to test new features and fixes as they are 'pushed' and passed through the CI.
Follow these download links for [Mac OS X 64 bit](https://packages.cloudfoundry.org/edge?arch=macosx64&source=github&version=v6), [Windows 64 bit](https://packages.cloudfoundry.org/edge?arch=windows64&source=github&version=v6) and [Linux 64 bit](https://packages.cloudfoundry.org/edge?arch=linux64&source=github&version=v6).
