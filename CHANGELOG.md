##v6.7.0
* Display correct information about app in copy-source -Restart app.Start/Stop/Restart/WatchStaging by passing org and
space name instead of assuming config contained correct information [finishes #81219748]

* Change initial output for copy-source [finishes #82171880]

* Add crypto/sha512 to import to solve unkown authority bug [Fixes #82254112]

* Fixes bug where null json value caused panic [Fixes #82292538]

* Merge pull request #290 from haydonryan/master Correcting status message

* Correcting status message previously space was set to org and vice versa, correcting.

* Fix french wording https://github.com/cloudfoundry/cli/pull/279 [finishes #81865644]

* Update application.PackageUpdatedAt to marshal json as time.Time [#82138922]

* Decolorize output for plugin to parse. [Finishes #82051672]

* Fix issue when making requests without a body [#79025838]

* move plugin cli invocations to a struct, which is passed into Run(...)

* Testing interval output printing - add PrintCapturingNoOutput to ui object to avoid using stdout in net
package tests
- make sure we rewrite entire string during interval output printing by
printing a long line of empty spaces [finish #79025838]

* Progress inidicated during uploads (push and create/update buildpack) [Finishes #79025838]

* Correcting status message previously space was set to org and vice versa, correcting.

* Terminal output can be silenced when invoke cli command from a plugin [#81867022]

* Add plugin_examples and README [finishes #78236438]

* Remove errant text from copy-source help output [Finishes #81813144]

* Exit 1 when a plugin exits nonzero or panics [#81633932]

* plugins have names defined by method

* `cf org` now displays space quotas. [Finishes #77390184]

* Merge pull request #280 from cloudfoundry/missing-service-instance-error-message update-service shows an error if the instance is missing and no plan is ...

* update-service shows an error if the instance is missing and no plan is provided

* Add `cf check-route` command [finishes #78473792]

* Plugins now have access to stdin (can be interactive) [finishes #81210182]

* Cli checks command shortname during plugin install - Cli also checks short names for commands when determining execution.
  Useful to prevent people from mucking with plugin configs by hand. [Finishes #80842550]

* Merge branch 'thecadams-honor-keepalive'
* Merge branch 'honor-keepalive' of github.com:thecadams/cli

* Improve error message return when refresh token has expired [finishes #78130846]

* Disable service access proprly queries for organization. [Finishes #80867298]

* plugns receive output from core cli commands

* Display most recent package uploaded time for cf app [finishes #78427862]

* Add CF_PLUGIN_HOME to help text output [finishes #81147420]

* Set MinVersion for ssl to TLS1, removing support for SSLV3 [#81218916]

* Add VCAP_APPLICATION to cf env output [finishes #78533524]

* Update `cf env` to grab booleans and integers. [Finishes #79059944]

* Implement update_service command [#76633662]

* Wait to output OK until app is started in start command

* Update help text for create-user-provided-service [finishes: #75171038]

* All arguments/flags are passed to plugin when plugin command invoked [finishes #78234552]

* Provide error when install_plugin plugin collides with other plugin -Update error message for collision with core cli command [finishes #79400494]

* Implement command `cf oauth-token` [Finishes #77587324]

* Use cached plugin config data instead of rpcing the plugin

* Cf help shows plugin info based on plugin_config [#78234404]

* update plugin config to store data for each command
* install handles conflicting commands
* validate plugin binary upon install

* Update `cf env APPNAME` to display running/staging env variables. - Refactor GetEnv api call to use counterfiter fake [Finishes #79059944]

* cf exit gracefully when i18n.T() is not initialized for configurations [Finishes #80759488]

##v6.6.2
* Bump version to 6.6.2
* Update usage text for install/uninstall-plugin [finishes #80770062][finishes #80701742]
* Move test setup into beforeEach of plan_builder_test
* Fix install_plugin usage text [finshes #80701742]
* security group commands show tip about changes requiring restart [Finishes #75375696]
* Remove unused scripts (moved for gocd) [#78508732]
* update correct fixture path in test code
* update transaltions for uninstall plugin description text
* stop translating commands, add missed translated strings
* Tar exectutables before uploading artifacts from gocd
* Update build-and-release-gocd tooling
* Potential fix for windows gocd timeout. 
* Fix for flakey tests in rpc package.
* Use 32 bit binary to get version when building installers
* Revert "Get version from 32bit binary, since the agent is 32bit" This reverts commit 8f7ff830b48f0926215adb60e8512e023e942ba5.
* Implemented plugins advertising their own name. - Name space with plugin name instead of binary name.
- Expose plugins directory as part of plugin configuration object
- Cli and plugins ping each other for availability. If the ping fails,
  they will stop the servers after 1 second. [Finishes #79964866]
* Refacto plugin/rpc to setup bidirectional communication [#79964866]
* Refactor install plugin to use counterfeiter fake. [#79964866]
* Plugin pings cf when it is ready to accept commands. - removes sleep from cf. [#79964866]
* refactor ServeCommand calls
* Change fake_word_generator to a counterfeiter fake [#74259334]
* add gi18n-checkup to bin/test [Finishes #80173376]
* Improve spacing for help output in create/update-space-quota [finishes #80052722]
* Add scripts for build-and-release for gocd
* Sync words.go with the word list [#80335818]
* Update error text on invalid json format. [Finishes #77391788]
* Improve help text for create-security-group command [Finishes #77391788]
* help will run as a core command instead of calling plugin commands [Finishes #78234942]
* plugin server runs on randomly chosen port
* consolodate plugin port configuration
* cf help includes plugin commands
* attempt to fix install paths for windows
* fix windows test failures by naming binaries with .exe extension
* close test file before deleting
* Fix error message for login w/ -a when ssl-cert invalid [#69644266]
* Finished refactor of configuration repository. [#78234942]
* Refactor plugin commands into rpc package -Also increase locales_test timeout
-Add empty_plugin executable to gitignore [#78234942]
* Refactoring plugins to include common code for rpc model. - plugins/rpc contains everything main used to contain.
- new interface for listing commands through rpc.
* Implement 'plugins' to list all installed plugin methods and the executable they belong to. [Finishes #78235118]
* go get godep before tests
* Revert "Use filepath instead of path where possible" This reverts commit 49beccf7726887211cfb05a20f6bbc175ec5847e.
- Failed on CI
* Use filepath instead of path where possible -Path does not always work well with windows [#79748230]
* Append .exe to config.json for plugin-config
* Name test binaries w/ .exe so windows WORKS
* Use filepath instead of path in main_suite_test -Add more debugging as well
* Add debugging statements to building plugin in main_suite_test
* Revert "Update GOPATH var in windows bat scripts" This reverts commit d311d8d4e71db7f8aad7d39d2ab0e1e26394aac2.
* Update GOPATH var in windows bat scripts
* Add debugging info to the main test
* Add ginkgo defer to allow us to see error message -This is when the main_suite_test fails before running
the main_test
* Skip checking filemode for instal-plugin on windows
* Retry request on tcp connection error. [Finishes #79151504]
* Added tests for the package main on windows during ci
* Added defaults for create-space-quota's help [Finishes #77394232]
* Improve testing with plugins and fix install-plugin bug -Chmod plugin binary after copying to the CF_HOME directory
-Test that all plugins work when multiple are successfully installed [finishes #78399080] [finishes #79487120]
* Refactor app instances to use a counterfeiter fake
* Fix tests relating to plugins and polution caused by them -Reduce sleep time when waiting for plugin to start
-Have main_test use plugin config the whole time in case of
invalid config in the home directory (the real home dir) [finishes #79305568]
* Wip commit for plugins with multiple commands
* Wip commit for plugins with multiple commands
* Add missing fixtures plugin command file.
* Compile test plugin every run. -This gives us a cross-platform test suite.
-Refactoring stuff out of main will make the test suite faster..
* Update changelog
* First pass at rpc model - have hardcoded port 20001
- sleep for 3 seconds waiting for rpc server [Finishes #78397654]

##v6.6.1
* Bump version to 6.6.1
* fix argument in callCoreCommand()
* Fix http_test.go to be OS independent [#79151902]
* Update flag descriptions for enable/disable service access [#79151902]
* show help when `cf` is input [#78233706]
Signed-off-by: Daniel Lavine <dlavine@us.ibm.com>
* Up tcp timeout to 10 seconds and log errors more effectively -Upping the timeout to deal with possible architecture issues, but
this should not be increased any more than 10 seconds
[#79151504]
* User can specify go program as a plugin in config.json [#78233706]
* Bump Goderps
* Dont pull from a locked SHA
* Lock CATS to a known good SHA (for now)
* Brought app_files repo into alignment with our new patterns. [#74259334]
* Revert "Update herd-cats-linux64 script to dynamically generate config" This reverts commit 7a74e5a3bfbb4e975eee4aedcc5a1471939070fc.
* Update herd-cats-linux64 script to dynamically generate config
* Move integration tests into main_test suite -Go 1.3 changes the way tests are built
* Move app_events repo into its own package. [#74259334]
* Upgrade to Go 1.3.1 - Go 1.3.x no longer orders maps, so we had to compensate in some of our
  tests.
- The fake server is a little smarter about "q" params now.
[Finishes #73583562]

* Bump Godeps for jibber-jabber. - Pull in Windows XP fix.

[Finishes #78489056]

* Remove -u option and clean up symlink in the build script.
* Bump Goderps
* Another attempt to fix unit tests on Windows
* Attempt to fix unit tests on Windows
* Change fake and refactor app_bits repo. - App bits repo is much more tightly scoped
- The App Bits repo has a counterfeiter fake, and lives in its own
  package
- Some callbacks met their demise
- We now have a push actor
- Former responsibilities of the App Bits repo have been divided between
  the App Bits repo, the push command, and the push actor.
- All this should make the future implementation of an "upload bits"
  command much easier/possible.
[#74259334]
* Change "-1" to "unlimited" in space-quotas. [#77830744]
* Change '-1' to 'unlimited' in space-quota. [#77830744]
* Display "unlimited" instead of "-1" in quota. [#77830744]
* Display "unlimited" instead of "-1" in quotas. [#77830744]
* Make Windows recognize PATH update and don't append on reinstall. [#78348074]
* Chmod the Inno Setup script. [#78348074]
* Change Windows installer build process to use Inno Setup. [#78348074]

## v6.6.0
* Modify set-running-environment-variable-group command usage to show example. [Finishes #77830856]
* Modify set-staging-environment-variable-group usage to show example of JSON. [Finishes #77837402]
* Add -i parameter for create-quota in usage. [Finishes #78111444]
* Can set locale using `cf config --locale LOCALE` - can clear locale providing CLEAR as flag argument. [Finishes #74651616]
* Implement set-running-environment-variable-group command. [Finishes #77830856]
* Implement "set-staging-environment-variable-group" command. [Finishes #77837402]
* Implement staging-environment-variable-group command. [Finishes #77829608]
* Implement running-environment-variable-group command. [Finishes #76840940]
* Make help for start timeouts on push more explicit. [Finishes #75071698]
* Implement disable-feature-flag command. [Finishes #77676754]
* Accept a bare -1 as instance memory on updating quotas. [#77765852]
* Implement enable-feature-flag command. [Finishes #77665980]
* Implement "feature-flag" command. Finishes #77222224]
* Can create organization with specified quota. [Finishes #75915142]
* Implement feature-flags command. [Finishes #77665936]
* Correctly accept a -1 value for creating quotas. [Fixes #77765852]
* Correctly display instance memory limit field for quotas. [Fixes #77765166]

## v6.5.1 
* Revert changes to update-service-broker. This cause a breaking change by mistake.

## v6.5.0
* Implement Space Quota commands (create, update, delete, list, assignment)
* Change cf space command to show information on the quota associated with the space. [#77389658]
* Tweak help text for "push" [#76417956]
* Remove default async timeout. [#76995182]
* Change update-service-broker to take in optional flags. [#63480754]
* Update plan visibility search to take advantage of API queries [#76753494]
* Add instance memory to quota, quotas, and update-quota. [#76292608]

## v6.4.0
* Implement service-access command.
* Implement enable-service-access command.
* Implement disable-service-access command.
* Merge pull request #237 from sykesm/hm-unknown-instances Use '?' instead of '-1' when running instances is unknown [#76461268]
* Merge pull request #239 from johannespetzold/loggregator-debug-printer CF_TRACE option for cf logs
* Stop using deprecated endpoints for domains. [#76723550]
* Refresh auth token on all service-access commands. [#76831670]
* Stop CLI from hanging when Loggregator keeps returning errors. [#76545800]
* Merge pull request #234 from fraenkel/cfignoreIgnored Copy cfignore to upload directory to properly ignore files
* Pass in ProxyFromEnvironment function to loggregator_consumer. [#75343416]
* Merge pull request #227 from XenoPhex/master By Grabthar hammer, by the sons of Worvan, you shall be avenged. Also, sorting.
* Add cli version to the "aww shucks" messsage. [#75131050]
* Merge pull request #223 from fraenkel:connectTimeout Use a connect timeout whenever making connections
* Merge pull request #225 from cloudfoundry/flush-log-messages Fix inter-woven output during start
* Merge pull request #222 from fraenkel/closeBody Close the response body
* Merge pull request #221 from jpalermo/master Fix base64 padding

## v6.3.2
* Provides "pretty printed" output of config JSON. [#74664516]
* Undo recursive copy of files [#75530934]
* Merge all translations into monolithic files. [#74408246]
* Remove some words from dictionary [#75469600]
* Merge pull request #210 from wdneto/pt_br Initial pt-br translation [#75083626]

## v6.3.1
* Remove Korean as a supported language. - goi18n does not currently support it, so it is in the same boat as Russian.
* Forcing default domain to be the first shared domain. Closes #209 [#75067850]
* The ru_RU locale is not supported. The go-i18n tool that we use does not support this locale at the moment and thus we should not be offering translation until such time as that changes. Closes #208 [#75021420]
* Adding in tool to fix json formatting
* Fixes spacing and file permissions for all JSON files. Spacing i/s now a standard 3 spaces. Permissions are now 0644.
* Merges Spanish Translations. Thanks, @bonzofenix! Merge pr/207 [#74857552]
* Merge Chinese Translations from a lot of effort by @wayneeseguin. Thanks also to @tsjsdbd, @isuperbb, @shenyefeng, @hujie5592427, @haojun, @wsxiaozhang and @Kaixiang! Closes #205 [#74772500]
* Travis-CI builds should run i18n tests Also, fail if any of those other commands fail

## v6.3.0
* Add commands for managing security groups
* Push no longer uses deprecated endpoint for domains. [#74737286]
* `cf` always returns exit code 1 on error [#74565136]
* Json is interpreted properly for create/update user-provided-service. Fixes issue #193 [#73971288]
* Made '--help' flag match the help text from the 'help' command [Finishes #73655496]

## v6.2.0
* Internationalize the CLI [#70551274](https://www.pivotaltracker.com/story/show/70551274), [#71441196](https://www.pivotaltracker.com/story/show/71441196), [#72633034](https://www.pivotaltracker.com/story/show/72633034), [#72633034](https://www.pivotaltracker.com/story/show/72633034), [#72633036](https://www.pivotaltracker.com/story/show/72633036), [#72633038](https://www.pivotaltracker.com/story/show/72633038), [#72633042](https://www.pivotaltracker.com/story/show/72633042), [#72633044](https://www.pivotaltracker.com/story/show/72633044), [#72633056](https://www.pivotaltracker.com/story/show/72633056), [#72633062](https://www.pivotaltracker.com/story/show/72633062), [#72633064](https://www.pivotaltracker.com/story/show/72633064), [#72633066](https://www.pivotaltracker.com/story/show/72633066), [#72633068](https://www.pivotaltracker.com/story/show/72633068), [#72633070](https://www.pivotaltracker.com/story/show/72633070), [#72633074](https://www.pivotaltracker.com/story/show/72633074), [#72633080](https://www.pivotaltracker.com/story/show/72633080), [#72633084](https://www.pivotaltracker.com/story/show/72633084), [#72633086](https://www.pivotaltracker.com/story/show/72633086), [#72633088](https://www.pivotaltracker.com/story/show/72633088), [#72633090](https://www.pivotaltracker.com/story/show/72633090), [#72633090](https://www.pivotaltracker.com/story/show/72633090), [#72633096](https://www.pivotaltracker.com/story/show/72633096), [#72633100](https://www.pivotaltracker.com/story/show/72633100), [#72633102](https://www.pivotaltracker.com/story/show/72633102), [#72633112](https://www.pivotaltracker.com/story/show/72633112), [#72633116](https://www.pivotaltracker.com/story/show/72633116), [#72633118](https://www.pivotaltracker.com/story/show/72633118), [#72633126](https://www.pivotaltracker.com/story/show/72633126), [#72633128](https://www.pivotaltracker.com/story/show/72633128), [#72633130](https://www.pivotaltracker.com/story/show/72633130), [#70551274](https://www.pivotaltracker.com/story/show/70551274), [#71347218](https://www.pivotaltracker.com/story/show/71347218), [#71441196](https://www.pivotaltracker.com/story/show/71441196), [#71594662](https://www.pivotaltracker.com/story/show/71594662), [#71801388](https://www.pivotaltracker.com/story/show/71801388), [#72250906](https://www.pivotaltracker.com/story/show/72250906), [#72543282](https://www.pivotaltracker.com/story/show/72543282), [#72543404](https://www.pivotaltracker.com/story/show/72543404), [#72543994](https://www.pivotaltracker.com/story/show/72543994), [#72548944](https://www.pivotaltracker.com/story/show/72548944), [#72633064](https://www.pivotaltracker.com/story/show/72633064), [#72633108](https://www.pivotaltracker.com/story/show/72633108), [#72663452](https://www.pivotaltracker.com/story/show/72663452), [#73216920](https://www.pivotaltracker.com/story/show/73216920), [#73351056](https://www.pivotaltracker.com/story/show/73351056), [#73351056](https://www.pivotaltracker.com/story/show/73351056)]
* 'purge-service-offering' should fail if the request fails [[#73009140](https://www.pivotaltracker.com/story/show/73009140)]
* Pretty print JSON for `cf curl` [[#71425006](https://www.pivotaltracker.com/story/show/71425006)]
* CURL output can be directed to file via parameter `--output`.  [[#72659362](https://www.pivotaltracker.com/story/show/72659362)]
* Fix a source of flakiness in start [[#71778246](https://www.pivotaltracker.com/story/show/71778246)]
* Add build date time to the `--version` message, `cf --version` now reports [ISO 8601](http://en.wikipedia.org/wiki/ISO_8601) date [[#71446932](https://www.pivotaltracker.com/story/show/71446932)]
* Show system environment variables with `cf env` [[#71250896](https://www.pivotaltracker.com/story/show/71250896)]
* Fix double confirm prompt bug [[#70960378](https://www.pivotaltracker.com/story/show/70960378)]
* Fix create-buildpack from local directory [[#70766292](https://www.pivotaltracker.com/story/show/70766292)]
* Gateway respects user-defined Async timeout [[#71039042](https://www.pivotaltracker.com/story/show/71039042)]
* Bump async timeout to 10 minutes [[#70242130](https://www.pivotaltracker.com/story/show/70242130)]
* Trace should also respect the user config setting [[#71045364](https://www.pivotaltracker.com/story/show/71045364)]
* Add a 'cf config' command [[#70242276](https://www.pivotaltracker.com/story/show/70242276)]
  - Uses --color value to enable/disable/ignore coloring [[#71045474](https://www.pivotaltracker.com/story/show/71045474), [#68903282](https://www.pivotaltracker.com/story/show/68903282)]
  - Add config --trace flag [[#68903318](https://www.pivotaltracker.com/story/show/68903318)]

## v6.1.2
* Added BUILDING.md document to describe our CI / build process
* Fixed regression where the last few log messages received would never be shown
  - affected commands include `cf start`, `cf logs` and `cf push`
* Fixed a bug in `cf push` related to windows and empty directories [#70470232] [#157](https://github.com/cloudfoundry/cli/issues/157)
* Fixed a bug in `cf space-users` and `cf org-users` that would incorrectly show all users
* `cf org $ORG_NAME` now displays the quota assigned to the org
* Fixed a bug where no log messages would be received if your access token had expired [#66242222]

## v6.1.1
- New quota CRUD commands for admins
- Only ignore `manifest.yml` at the app root directory [#70044992]
- Updating loggregator library experimental support for proxies [#70022322]
- Provide a `--sso` flag to `cf login` for SAML [#69963402, #69963432]
- Do not use deprecated domain endpoints in `cf push` [#69827262]
- Display `X-Cf-Warnings` at the end of all commands [#69300730]
* Add an `actor` column to the `cf events` table [#68771710]

## v6.1.0
* Refresh auth token at the beginning of `cf push` [#69034628]
* `cf routes` should have an org and space requirement [#68917070]
* Fix a bug with binding services in manifests [#68768046]
* Make delete confirmation messages more consistent [#62852994]
* Don`t upload manifest.yml by default [#68952284]
* Ignore mercurial metadata from app upload [#68952326]
* Make delete commands output more consistent [#62283088]
* Make `cf create-user` idempotent [#67241604]
* Allow `cf unset-env` to remove the last env var an app has [#68879028]
* Add a datetime for when the binary was built [#68515588]
* Omit application files when CC reports all files are staged [#68290696]
* Show actual error message from server on async job failure [#65222140]
* Use new domains endpoints based on API version [#64525814]
* Use different events APIs based on API version [#64525814]
* Updated help text and messaging
* Events commands only shows last 50 events in reverse chronological order [#67248400, #63488318, #66900178]
* Add -r flag to `cf delete` for deleting all the routes mapped to the app [#65781990]
* Scope route listed to the current space [#59926924]
* Include empty directories when pushing apps [#63163454]
* Fetch UAA endpoint in auth command [#68035332]
* Improve error message when memory/disk is given w/o unit [#64359068]
* Only allow positive instances, memory or disk for `cf push` and `cf scale` [#66799710]
* Allow passing "null" as a buildpack url for "cf push" [#67054262]
* Add disk quota flag to push cmd [#65444560]
* Add a script for updating links to stable release [#67993678]
* Suggest using random-route when route is already taken [#66791058]
* Prompt user for all password-type credentials in login [#67864534]
* Add random-route property to manifests (push treats this the same as the --random-hostname flag) [#62086514]
* Add --random-route flag to `cf push` [#62086514]
* Fix create-user when UAA is being directly used as auth server (if the authorization server doesn`t return an UAA endpoint link, assume that the auth server is the UAA, and use it for user management) [#67477014]
* `cf create-user` hides private data in `CF_TRACE` [#67055200]
* Persist SSLDisabled flag on config [#66528632]
* Respect --skip-ssl-validation flag [#66528632]
* Hide passwords in `CF_TRACE` [#67055218]
* Improve `cf api` and `cf login` error message around SSL validation errors [#67048868]
* In `cf api`, fail if protocol not specified and ssl cert invalid [#67048868]
* Clear session at beginning of `cf auth` [#66638776]
* When renaming targetted org, update org name in config file [#63087464]
* Make `cf target` clear org and space when necessary [#66713898]
* Add a -f flag to scale to force [#64067896]
* Add a confirmation prompt to `cf scale` [#64067896]
* Verify SSL certs when fetching buildpacks [#66365558]
* OS X installer errors out when attempting to install on pre 10.7 [#66547206]
* Add ability to scale app`s disk limit [#65444078]
* Switch out Gamble for candied yaml [#66181944]

## v6.0.2
* Fixed `cf push -p path/to/app.zip` on windows with zip files (eg: .zip, .war, .jar)

## v6.0.1
* Added purge-service-offering and migrate-service-instances commands
* Added -a flag to `cf org-users` that makes the command display all users, rather than only privileged users (#46)
* Fixed a bug when manifest.yml was zero bytes
* Improved error messages for commands that reference users (#79)
* Fixed crash when a manifest didn`t contain environment variables but there were environment variables set for the app previously
* Improved error messages for commands that require an API endpoint to be set
* Added timeout to all asynchronous requests
* Fixed `bad file descriptor` crash when API token expired before file upload
* Added timestamps and version information to request logs when `CF_TRACE` is enabled
* Added fallback to default log server endpoint for compatibility with older CF deployments
* Improved error messages for services and target commands
* Added support for URLs as arguments to create-buildpack command
* Added a homebrew recipe for cf -- usage: brew install cloudfoundry-cli
