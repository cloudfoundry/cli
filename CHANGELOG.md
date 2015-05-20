##v6.11.3
* Improve Tip for bind-service command [#94153632]
* fix bug where app's PackageState is incorrectly set in restage [#93382608]
* Merge branch 'hwcf-issue-32' of https://github.com/HuaweiTech/cli into HuaweiTech-hwcf-issue-32
* fixed push -p help verbiage
* refactor to make err will always be caught in start.go
* improve error checking after calling endpoint [#93382608]
* use proper model for /apps endpoint [#93382608]
* using /apps instead /instances to poll for staging [#93382608]
* Translate failure message for invalid JSON in arbitary params arg for create-service [#88670540]
* Add French translation for arbitrary params description
* new staging_failed_reason field in App Model [#93382608]
* new GetApp() method in ApplicationRepository package [#93382608]
* add package_state to App Model [#93382608]
* fix conflicts in language files 
* do not create zip when no file to upload [#94014700]
* updated and resolves conflicts in language files [#94014700]
* Add -c flag to pass arbitrary params during create-service [#89843658]
* Remove async from request body during create-service Two problems here:[#92396108]
  1. Async is a query parameter flag, not a post body paramter
  2. POST /v2/service_instances does not respect async flag anyway
* Merge pull request #427 from xingzhou/service_key_delete add delete service key command
* cf start uses old loggregator to tail logs, instad of noaa [#93554176]
* use old loggregator consumer to retrieve logs [#93554176]
* godeps [#93554176]
* add old loggregator_consumer package [#93554176]
* rename noaa specific packages [#93554176]
* Merge pull request #415 from HuaweiTech/hwcf-issue-30  Fix for stack and stacks command
* add delete service key command [#87062548]
* Fix for stack and stacks command

##v6.11.2
* not renewing noaa consumer on every push instead, we instruct noaa to stop reconnecting in the background
* hardcode doppler endpoints into config getter [#93208696]
* Fix for stack and stacks command
* Merge pull request #419 from xingzhou/service_key_get add show service key detail
* add show service key detail [#87061876]
* Merge pull request #396 from xingzhou/service_key_list added service keys command
* minor fixes for max's comments on service key list PR [#87057920]

##v6.11.1
* close channel properly during re-auth when connecting with noaa [#92716720]
* 20 second timeout for connecting to logging server while pushing [#92702342]
* mutex to avoid race condition [#92702342]
* renew the noaa obj when pushing mutilple apps to avoid stalling bug [#92716720]
* enable re-instantiating noaa obj in app starter [#92716720]
* deps noaa package
* added service keys command [#87057920]
* fix panicing when slice contains invalid values [#92135482]
* Updated gi18n binary name

##v6.11.0
* Fixed more version checking tests 
* Fixed version check tests 
* Changed update message to min-cli-version, not min-reccommended-version 
* Updated translation files. Removed duplicate entries in translation files. 
* Added version checking to login. Finishes [#92086178]
* Updated gi18n package name in bin/gi18n-checkup 
* `cf target` now checks for minimum CLI version. [Finishes #92086308]
* login command prompts user to update cli version [finishes #86074346]
* get min_cli_version from CC [#86074346]
* Merge pull request #400 from att-cloudfoundry/rd7869-patch-1 Update README.md
* associate stack with an app in `cf app` [finishes #91056294]
* Merge pull request #397 from xingzhou/service_key Print the "not authorized" error returned from CC when creating service key
* Added Min CLI and Reecommended CLI version numbers to config. [Finishes #86074256]
* Print out the "not authorized" error returned from CC when creating service keys Fix a bug that only the spacedeveloper or admin can create a service key. CC will return "Not authorized error" and CLI need to report the error and print out the error message.
* Merge pull request #385 from xingzhou/service_key Add 'create-service-key' command in cli [#87057732, #87157018]
* Merge pull request #384 from cloudfoundry/async Show blank last operation if the CC returns null last_operation in API response.
* fix bug in logging unit test
* improve error reporting during log tailing Signed-off-by: Jonathan Berkhahn <jaberkha@us.ibm.com>
* Merge pull request #375 from HuaweiTech/hwcf-issue-22 Updated the package path
* avoid closing channel twice 
* quit listening loop properly while tailing logs
* go fmt
* godeps - remove loggregator_consumer [finishes #83692758]
* use noaa to tail logs/get recent logs [#83692758]
* use noaa instead of loggregator_consumer when getting recent logs [#83692758]
* Add 'create-service-key' command in cli 1. Add a new command named "create-service-key" to create a service key
for a specified service instance.
2. Add error of unbindable service
[finishes #87057732 & #87157018]
* enable bool flag value to be set
* populate Args() and accept form in '-flag=value' [finsihes #90067220]
* flag parsing: int, bool, string [#90067220]
* allows multiple domains in app manifest [finishes #88801706]
* add domains field to manifest [#88801706]
* update help text: buildpack 'null/default' usage [finishes #89827178]
* language files for command cups help [#90319606]
* windows help example for command cups [finishes #90319606]
* return correct error when unable to create config [finishes #88666504]
* manifest.yml now supports `no-hostname` field [finsihes #88386830]
* Update README.md 
* bump candiedyaml version [finishes #89305904]
* improve help text examples for `cf login` [finishes #89650282]
* Merge pull request #379 from HuaweiTech/hwcf-issue-17 Added way to put user in org's space with 'cf target -o ORG' command if there is only one space
* Merge pull request #344 from HuaweiTech/hwcf-issue-9 Adding a way to see Security Group Rules
* Added way to put user in org's space with 'cf target -o ORG' command if there is only one space cf target with [-o] flag will internally target org's space if there is only one space. [#73568408]
* Merge pull request #353 from fraenkel/shared_private_domains Shared private domains
* better error message when tmp dir does not exist while not load language files [finishes #86888672]
* --guid flag for command stack [finishes #89221186]
* new command `stack` [finishes #89220886]
* Update README.md 
* Merge pull request #360 from SrinivasChilveri/hwcf-issue-11 Fix the requirmements issue in some of the application commands
* Make OrgReq and SpaceReq creation concurrency-safe for plugins. [Finishes #89473078]

* Updated the package path 
* fixes error when plugin rpc server is not reachable
* closes client rpc connection [finishes #89307102]
* Merge pull request #345 from simonleung8/master Ginkgo matcher BeInDisplayOrder()
* godeps
* `app` command gets metric directly from loggregator for diego app [finishes #89468688]
* noaa api library for diego app metric and fakes [#89468688]
* wrapper for noaa and fakes for tests [#89468688]
* comment explains temp solution for doppler endpoint [#89468688]
* add diego flag to app model [#89468688]
* read doppler endpoint from manifest [#89468688]
* populate doppler endpoint from loggreator endpoint [#89468688]
* fixes problem with plugin calling CLI concurrently - fixes ApplicationRequirement 404 error [finishes #89452922]
* Revert "closes http.Response body" This reverts commit 86a2b55bc1850369f500dd94ef2abb1998b4747a.
* closes http.Response body
* uses app.guid within route object to unmap routes [finishes #87160260]
* Merge pull request #363 from cloudfoundry/old_cc_update_plan_bug Prevent updating service plans when the CC is less than v191.
* Merge pull request #357 from cloudfoundry/async Changed service instance commands to yellow (CommandColor).
* Merge branch 'async' into old_cc_update_plan_bug 
* Remove unused import 
* fix bug where uninstall-plugin fails
* Prevent updating service plans when the CC is less than v191. v191 corresponds to CC API 2.16.0.
This is to prevent a bug with older CC and newer CLIs where plans can be
updated without talking to the service broker.
[#88798806, #88689444]
* update test fixtures to react to plugin uninstall 
* closing a file in test
* Plugin can call CoreCliCommands upon uninstalling - extract rpcService constructing into main
- pass rpcService to command_factory
- rpcService is passed into `install-plugin`, `uninstall-plugin`
[#88259326]
* made further reading into a bulleted list 
* Added plugin dev guide link to Further Reading section.  Now it appears in main readme twice 
* Made link to plugin docs **bold** 
* Update README.md 
* send `CLI-MESSAGE-UNINSTALL` to plugin upon uninstalling [finishes #88259326]
* Fixed OK message formatting in enable-service-access. [Finishes #86670482]
* Fix the requirmements issue in some of the application commands
* Changed service instance commands to yellow (CommandColor). [Fixes #86668046]
* Merge pull request #351 from cloudfoundry/async Finishes async work for CLI
* bubble up any error when zipping up files during push [#87228574]
* Added accepts_incomplete=true param to delete service instance. [#87584124].
* Updated text output when deleting services instances asynchronously. [Finishes #88279874]
* Updated text output when updating services instances asynchronously. [Finishes #88279828].
* Updated text output of cf create-service. [Finishes #86668046]
* Merge pull request #348 from SrinivasChilveri/hwcf-issue-2 Fix 'cf routes'output should be scoped to org and grouped by space
* Add new share/unshare private domains command - Allow an admin to share a private domain with an org
- Allow an admin to unshare a private domain with an org
* Detect private domains properly - Shared private domains make the owning org null
  Rather than check if owning_organization is present, check for the
  presence of the shared_organization_url
* Update CHANGELOG.md 
* Update README.md 
* Fix 'cf routes'output should be scoped to org and grouped by space Solution to the bug:- [#70300846]
* `service-brokers` uses BeInDisplayOrder() to assert output order 
* ginkgo matcher to assert string output order 
* Adding a way to see Security Group Rules

##v6.10.0
* rename default plugin repo
* Update README.md 
* Merge pull request #349 
* Added accepts_incomplete parameter to update and rename service. [#86584082]
* changed the async provisioning messages [#86668046]
* Update service instance last operation state => status 
* Formatting for services and service command matching new fashion [#86585678]
* changes commands for last_operation 'fashion' * create-service
* service
* services
* service-summary
* utils object constructor returns a pointer 
* `install-plugin` only tries downloading with internet prefixes 
* validate sha1 when installing plugin from repo [#86072988]
* utils for sha1 computing, comparing [#86072988]
* Changed list-plugin-repo to list-plugin-repos [Finishes #87851674]
* not asserting checksum in util test 
* take out checksum in assertion [#87856234]
* --checksum flag for command plugins [#87856234]
* sha1 checksum utils [#87856234]
* repo name case insensitive when installing plugins
* Plugin Repo default - plugins.cloudfoundry.org
* Godeps clipr
* not locating plugin binary locally if path prefix with internet address
* `list-plugin-repo` command [#86071226]
* trim internet addr prefix before checking file existance [#86073134]
* improve help text for command repo-plugins [finishes #86071226]
* `install-plugin` can install from a repository [#86073134]
* update file downloader [#86073134]
* Extract list plugins from repo functions into actors [#86073134]
* fix bug where args is overwritten itself before flags in testhelpers
* Repo name comparisons in add-plugin-repo are case-insensitive. [#87467254]
* Merge pull request #343 from fraenkel/instance_details
* App instance may contain additional details [#86856252]
* `repo-plugins` can list a plugins from a single repo with `-r` [#86071226]
* Added remove-plugin-repo command to remove plugin repos. [#86141272]
* new command `repo-plugins` - list plugins from all repos [#86071636]
* `cf service-brokers` output sorted by name [#86663258]
* remove commented code 
* CLI knows about 'CRASHED' in addition to 'FLAPPING' [#87141282]
* Godeps clipr 
* new `add-plugin-repo` commnad [#86452004]
* improved plugin topics for help [#86452004]
* config Getter & Setter for PluginRepos [#86452004]
* new PluginRepos field in config.json [#86452004]
* Removed help references to specific companies. [#87059156]
* non admin can see other users with `space-users` [#86963130]
* update fakes for user_repo [#86963130]
* new func to list space users w/o hitting UAA with api version >v2.21.0 [#86963130]
* non-privileged users can list users with `org-users` [#82059018]
* Add CallCount in fakes for testing [#82059018]
* Add Api version comparing to config [#82059018]
* new func to list org users w/o hitting UAA with api version >v2.21.0 [#82059018]
* Merge pull request #339 from cloudfoundry/async Async Service Provisioning
* Fixed bug where `cf services` would not parse the JSON [#62068908]
* Changing expected state from CC to be: * `in progress` vs `creating`
* `succeeded` vs `created` [#86578718]
* Changes text to user for status to be: * create succeeded
* create failed [#86578582]
* Notify user manifest is not found on `cf push` [#86561070]
* `create-app-manifest` now named the file <app-name>_manifest.yml [#86561764]
* Update README.md 

##v6.9.0 
* Merge PR #333: CLI sends async request for service instance provisioning
* Revert "new command user-provided-services" [#79188196]
* cf service(s) emits 'available' for services that do have a state. [#86181724]
* Renamed accept_unavailable to accepts_incomplete. [#86259450]
* Fixed table and detail formatting for service instances. [#62068908]
* changed NA to "" string for user provided service [#84252876]
* changed $cf service to add Status|Operation|Message sections [#84252876]
* added fixed status and (operation) for $cf services command [#84252876] 
* added check for ServiceInstance.State in CreateService [#62068908]
* Add State and StateDescription to service_instance [#62068908]
* Adding accept_unavailable=true query param for create-service [#62068908]
* new command user-provided-services [#79188196]
* counterfeiter fake for user_provided_service [#79188196]
* new GetSummaries() in api/user_provided_server.go [#79188196]
* fix usage of test http server [#79188196]
* new models: user-provided-service [#79188196]
* Correct help text for `files` command [#85754150]
* clarify comment for usage of TotalArgs
* Improve cf <commands> usage instructions [#85818652]
* Merge PR #328 from Fix cups attempts to create service when no space is targeted
* append source index to all source [#85484012]
* Update README.md add link to plugin development guide
* Update README.md Added link to complete plugin change log.
* Update Plugin CHANGELOG.md Changed CHANGELOG.md to complete list of all plugin feature changes.
* Update Plugin CHANGELOG.md Added version 6.7.0 info.
* Update Plugin README.md Added version 6.8.0 info.
* Touch change log for example plugins.
* includes [HEALTH/{index}] from diego log [#85484012]
* Merge PR #322: Updating go vet location in install-dev-tools target.
* Merge PR #323: Fixes go vet errors:
* Usage help example for plugins [#85665592]
* remove '-' in test_1 plugin help sample
* Merge PR #321: Copy original request's headers when handling redirect
* Fix attempts to create service even when no space is targeted Solution to the bug [#82753668]
* improvement to marketplace cost messaging [#85571986]
* Update plugin example readme 
* Additional readme for plugin/rpc workflow 
* addition diagram for plugin rpc workflow
* Update README to detail plugin/cli interaction 
* illustrative diagram for plugin example README 
* update TestCommandFactory for new interface
* main refactor, extract code into command_factory New func in command_factory
* GetByCmdName() can finds by short name [#82051134]
* enable plugin commands to allow '-h' and '--help' flags [#82051134]
* merge plugin metas and core command metas to be used in codegangster [#82051134]
* extract getting plugin metadata out of RunMethodIfExists() [#82051134]
* Add usage to test plugins and set version numbers to be different [#82051134]
* Plugin usage/option model, for use in help [#82051134]
* Fixes go vet errors
* Updating go vet location in install-dev-tools target
* Update README in plugin example for versioning [#85484250]
* plugin example to show versioning usage [#85484250]
* Copy original request's headers when handling redirect (fixes #318 on github)
* `cf plugins` shows plugin versions [#84630868]
* write version to config when install plugin [#82911038]
* Allow versioning in plugins [#82911038]
* Merge PR #317: Fix the invalid memory address during bind service
* document new buildpack specifiers feature [#75205334]
* Merge PR #315: Improve french i18n
* Fix the invalid memory address during bind service Solution to the bug [#79267756]
* fixed spelling in changelog.md [#84867042]
* Merge PR #309: Fix in clearing space fields of config data on cf space-delete
* Better message when no files to be listed in directory [#63120324]
* Allows both host and hosts in manifest [#72389932]
* allows multiple hosts(routes) to be created when app is pushed [#72389932]
* Add hosts field for manifests [#72389932]
* Preserve user-provided vars type when generating manifest. [#78294704]
* Sort Environment Vars in manifest alphabetically [#78294704]
* Includes startup command in `create-app-manifest` [#78294704]
* New Command field in generated manifest [#78294704]
* Apps now timeout when they fail to stage insead of waiting for an instance to start [#83802536]
* i18n for install-plugin help text
* improve help text for install-plugin [#84601290]
* skip validating negative integer when it is a value to another flag [#84317640]
* skip flag verification for arguments, only verify flags [#84317640]
* replace file.Write() with fmt.Fprintf() in generate_manifest.go 
* remove unused func in generate_manifest.go 
* fix generated mainfest formet from create-app-manifest [#78294704]
* command create-app-manifest for generating manifest for pushed app [#78294704]
* new func to assert manifest orders in test [#78294704]
* new package for generating manifests [#78294704]
* fake for generate_manifest.go [#78294704]
* add health_check_timeout to Application model [#78294704]
* populates EnvironmentVars when hitting app/summary endpoint [#78294704]
* Add services to models.Application [#78294704]
* remove unsed code in mainfest.go 
* Fix in clearing space fields of config data on cf space-delete 

##v6.8.0
* Allows plugin to be installed from an Url [#80043644]
* Allows mutliple plugins with blank aliases. [#84241752]
* Remove commented line in update_service_test 
* test fix and additional coverage [#80043644]
* Exit non-zero in build-and-release-gocd if sub-script fails
* New utils for download single file from url 
* create-buildpack and update-buildpack now allow relative paths. [#80043644]
* Update ginkgo
* Add `cf restart-app-instance` command [#78049908]
* Add dashboard-url to `cf service` output [#68396596]
* Add unset flag to `cf api` -Allows user to unset the api endpoint [#82979408]
* `cf plugins` shows command alias [finishes #83892154]
* plugin alias shows in `cf help` [finishes #83892240]
* improve error text for plugin alias conflict errors. [#83717740]
* `cf install-plugin` cross-checks for command/alias conflicts [#83717740]
* Fixed plugin test fixture; Made aliases work with multi-command plugins 
* Added aliases for plugins. [#82051186]
* README update for multi-command plugin example [#83690584]
* code example for plugin with multiple commands [#83690584]
* improve text in help [#82913246]
* correct display order in space admin help section [#83437508]
* `cf org` displays all information in quota [#83363414]
* improve help text for command `uups` [#83233266]
* Add guid flag to `cf org` [#83435546]
* Add guid flag to `cf space` [#83435684]
* Add guid flag to `cf service` [#83435846]
* Update README.md 
* fake out cf config for testing [#82871316]
* Merge branch 'hw-issue-20' of github.com:HuaweiTech/cli into HuaweiTech-hw-issue-20 
* Merge branch 'hw-issue-21' of github.com:HuaweiTech/cli into HuaweiTech-hw-issue-21 
* Update buildpack flag descriptions [#83069682]
* Allow users to specify a space-quota when creating a space [#82311654]
* Update travis golang version to reflect the version we compile on
* Attempt to fix travis build with ginkgo flag [#82012788]
* Update ginkgo 
* Show detected_start_command on first push [#79325064]
* Merge pull request #287 from HuaweiTech/hw-issue-2 Extraneous arguments now cause commands to fail with usage.
* Prompt is always shown to user, even when the plugin has invoked the cli command with output suppressed. [#82770766]
* Update jibber_jabber - Adds support for zh-TW and has fix that moves zh-CK to zh-HK [#83146574]
* Merge pull request #299 from uzzz/master Fix ui.Ask to return strings with spaces from stdin
* Changed iscc to use environment variable for finding WINE.
* Replace hard coded path to restore the build and release script.
* Fix ui.Ask to return strings with spaces from stdin [#78596198] 
* Fix windows init_i18n test -Also fix compilation issues related to injection of jibber_jabber
* Inject jibberjabber so it can be tested Attempt to fix windws Hant/Hans init tests
* Revert "Revert "fix failing HK/TW Windows 32 unit test"" 
* Revert "Revert "Match traditional Chinese dialects to zh_Hant"" 
* Revert "Revert "Moved chinese translations to more generic locale tags"" 
* polling respects api target host while performing http 'Create' request [#77846300]
* polling respects api target host while performing http 'Update' request [#77846300]
* polling respects api target host while performing http 'Delete' request [#77846300]
* When starting an app the start command is displayed to the user [#79325064]
* Use '$HOME' env var instead of hard coded path 
* Use iscc in scripts directory when building installers
* Add comments to build-installers-gocd script for installation of 'Inno Setup 5'
* Add iscc file for creating windows installer
* Fix quota creation to default to unlimited instance memory [#82914568]
* Allow users to set quotas and space-quotas instance memory to 0 [#82914568]
* Fix the args validation in commands 
* Update help text for `cf update-buildpack` and `cf create-buildpack` [#82828946]
* Update README.md 
* Add command help text to `cf plugins` [#82777012]
* `-h` and `--h` should not report as invalid flags [#69038672]
* Add `--guid` flag to `cf app` - Allow users to get the guid of an application with a guid flag [#76459212]
* find plugins in the current directory without having to specify `./` [#82776732]
* Fix the usage info in cf feature-flag command 
* var renaming for readability 
* handles both "-" & "--" prefix for flag checking - ignores flag value after `=` [#69038672]
* T() up new texts for translation - dot-import i18n
* informs user about incorrect flags 
* Improve messaging `cf unmap-route` output [#82187142]
* Removing api requirement for `cf service-access` [#77468074]
* Revert "Moved chinese translations to more generic locale tags"
* Revert "Match traditional Chinese dialects to zh_Hant"
* Revert "fix failing HK/TW Windows 32 unit test"
* Fix the Usage info in cf security-groups command 
* fix failing HK/TW Windows 32 unit test 
* tip text for update-buildpack [#82910350]
* Merge pull request #297 from jberkhahn/default_english Match traditional Chinese dialects to zh_Hant
* Match traditional Chinese dialects to zh_Hant 
* update readme add step for running godep restore to ensure appropriate go dependencies are present
* Remove 'CommandDidPassRequirements' global test var [#70595398]
* 'service-access' command requires cc api version 2.13.0 
* Do not prompt the user for org when none are available during login [#78057906]
* Do not prompt the user for a space during login when the user has no available spaces [#78057906]
* Handle non 403 error when accessing the uaa endpoint 
* Add tip to `cf m` about the -s flag [#82673112]
* Update push --no-route help text to be more accurate [#64863370]
* Improve error handling for create-user [#80042984]
* Handle non string env var variables. 
* Moved chinese translations to more generic locale tags 
* Fix issue with create-service
* Update README.md 
* Update README.md 
* Merge pull request #293 from jennjblack/edits edit cf CLI dev guide README
* edit cli README.md 
* Update README.md Add Releases info to Download section of the README [#78473546]
* Show whether a service is paid in `cf m` [#76373558]
* Add script to improve release cutting process [#79626744]
* edit cli/plugin_examples README.md 
* Remove inline-relations-depth calls from service_builder calls [#81535612]
* `cf m -s service-name` works when unauthenticated [#81535612]
* Begin adding -s flag to `cf m` [#81535612]
* Update output for bad memory or disk quota in manifest [#79727218]
* Handle manifest memory and disk values that are numeric and have no memory unit [#79727218]
* Update output for bad memory or disk quota in manifest [#79727218]
* Handle manifest memory and disk values that are numeric and have no memory unit [#79727218]
* Improve 'cf unset-org-role' error message on Access Denied (code 403) [#77158010]
* User is warned when creating a service that incurs cost 
* edit cf CLI dev guide README 

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
