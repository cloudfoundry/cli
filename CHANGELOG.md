## 6.23.0
* Bump version to 6.23.0
* increase the dial timeout
* the verbose integration test should be testing against a redacted request body
* fix flakyness
* properly display non-json request body in the loggers
* Converts all everinoment variables in manifest to string
* renamed push_test to push_command_test
* display with flavor
* save refresh token when refreshing auth token
* added the logo
* remove experimental flag from several commands
* update go-interact and dependancy
* switching to using a custom DNS record instead of xip.io
* display request parameters when logging
* display extra tip with client errors for V3 commands
* skipping in windows
* test SSLValidationHostnameError using xip.io
* integration tests for delete-user cleans up after itself
* Unset and reset environment variables during config unit tests
* `delete-user` errors when deleting a non-unique user belonging to multiple origins
* fix translation
* always have CF_TRACE override config
* added godocs comments to API layer
* increasing dial timeout due to test flakiness
* get a better error message from x509.HostnameError
* run with experimental turned on
* bold the request and response headers in trace output
* added translations to updated `tasks` and `run-tasks` --help message
* `tasks` and `run-tasks` "see-also" message recommends logs
* add examples to create-user usage
* set async to true for the delete
* test display output in the correct buffer
* consistent command names in integration tests
* refactor delete-org command
* v3 commands verify minimum API version
* add postgres directory to diego-sql.yml path
* add sql stubs to deploy-diego task
* add app name/version to UAA client
* add integration test for delete org
* do not . import helpers
* test the experimental code
* create-user has support for --origin flag
* do not escape HTML when printing JSON body
* return the correct error when go-interact errors
* wrap UAA errors for v3 commands
* add logging to the UAA client for v3 commands
* output external IP in CATs tests
* switch to display warning to avoid printing 'FAILED'
* update color dependancies
* update proto buff, web sockets and ssh related packages
* update golang extensions packages
* update code.cloudfoundry.org vendored packages
* use cf-release-repo as gopath, use gvt on windows
* update ginkgo to support the -flakeAttempts flag
* update ginkgo and gomega
* move final cmd tests into plugins
* adding -flakeAttempts flag to integration tests
* remove GATs, it is no more it has ceased to be
* moved GATs plugin tests into integration
* added a common prefix for all users created in integration tests
* remove application gats
* clean up terminal logger output
* refresh token tests check std.err for warnings if `CF_CLI_EXPERIMENTAL` is true
* clean up user created in integration tests
* added integration tests that verify refactored create-user command behaves correctly
* add the create-user command to the refactor branch
* add https:// to CF_API in windows integration test
* the CATs version with the tar fix is not in cf-release master yet
* this suite file was useless, so I'm removing it.
* only call startWait.Done once in total
* global tests have their own package
* rename suite file
* update paths
* move tests into 'isolated' package
* move helpers from suite to helpers package
* adds the create-user command integration tests
* stop skippin'
* add cleanup as prep for running tests
* xargs is dumb
* clean up correctly, Anand screwed up
* don't pipefail
* 5 was too many
* prevents a negative wait group counter from occurring
* cloudcontroller and uaa connections reset previous http responses on retries and log correct response state
* additional cleanup, just in case
* don't display error twice, doing the hacky thing
* authentication middleware resets original http body in retries
* move version into common
* rewrite some cmd unit tests as integration tests
* set KeepAlive to 30 seconds in all Dialers
* godoc warning people not to import these packages
* refactored version command
* git hook to prevent pushing commits with anonymous Pivotal user
* disable broken character encoding cats
* integration test for custom oauth client id and secret
* set OAuth client and secret defaults in NewData
* set OAuth client and secret defaults in LoadConfig
* remove SetCFOAuthClient and SetCFOAuthClientSecret
* move help and command list into common package
* rename packages {v2,v3}/common -> {v2,v3}/shared
* rename package command/flags -> command/flag
* skipping these CATs tests since they don't actually work
* share common code from commands/v2 and commands/v3
* switch back to master, looks like the commit has been merged
* rename commands to command
* skipping test as it hasn't actually worked in a long time
* Fix linux cat SKIPS variable setting
* pend service key test from running
* skip tasks tests for now, since they're borked
* we don't need to run these tests
* spaces are dumb
* run all the tests?
* find the default shared domain directly instead of using API URL
* use random names for apps
* ensure cleanup always occurs, also reduce parallelism
* enables use of custom client id and secret when authenticating with UAA
* moved most of the cli/cli-acceptance-tests over to integration
* remove default CF OAuth Client
* update ports to be comma separated in security group help text
* default CFOAuthClient to "cf" if empty
* Merge branch 's-matyukevich-cf-oauth-client2'
* Merge branch 'cf-oauth-client' of git://github.com/s-matyukevich/cli into s-matyukevich-cf-oauth-client2
* remove cli-private resource
* add push test to integration suite
* rotate homebrew tap github key
* remove cli-private resource
* remove unused sign-windows-binary task
* remove dependency on the cli-private repo
* use the develop branch of cf-release for cf-acceptance-tests
* fix cli-integration resource uri
* change public git resource URLs to https
* created a read only org that is shared across tests
* fix description on update-security-group-command
* use more of the default eventually timeout
* cleanup task for pipeline integration tests
* add description to update-security-group help json
* move integration test cleanup to a script
* use the same org name prefix
* add description to create-security-group help json
* handle int parsing if passed in string
* Bump NOAA
* clean up all the orgs - integration tests
* files are created in a tmp dir and cleaned up properly
* why not just set the timeout higher?
* does not retry on any POST requests
* user-agent only contains the base command name
* catch ErrMarshal in main to print help
* update the flags parse integer error to be more user friendly
* read verbose from config now, tell your friends.
* display redacted authorization header in trace output
* use default domain for service integration tests
* skip portion of test because the test setup is incorrect
* service-key command exits with 1 when not found
* fix incorrect method call
* refactor utils/ui package
* move task commands under APPS in help -a output
* add --name option to run-task command
* add task commands to help -a
* added the terminate-task command
* change retry request logic to only trigger for 500
* renamed ccv3 RunTask to NewTask per client naming convention
* add message when targetting api without being logged in
* update integration tests
* handle errors for API endpoint mismatches
* don't display extra target information in API cmd
* align table display in API command
* change the error message when the app is not staged
* pretty prints the output from the api command
* return a friendlier error when creating tasks in a diegoless environment
* add cloudfoundry to windows worker path
* pass the original body on every http request retry
* update translations for the RunTaskError
* remove focus
* test empty tables properly for v3 tasks command
* task running related errors are now displayed cleanly
* windows has a slightly different message, so don't check the whole thing
* task actor no longer returns TasksNotFoundError
* update expected tables in integration tests for better match
* optional protocal cuz of windows
* add scheme to default apiURL for integration tests
* changed name as it seems to work in preview but not after commit
* added slack badge
* added integration tests for v3 tasks command
* added additional behavior to v3 tasks command
* do not hardcode api URL in proxy integration tests
* integration tests for verbose flag and proxy
* added CF logo
* add table display to tasks command
* added the v3 tasks command
* move common errors into cloudcontroller package
* display full stack trace for panics
* uaa client will adhere to the CF_DIAL_TIMEOUT
* forgot that API won't have the config set, this should fix it
* embed some structures and add some documentation
* pass user-agent and connection headers to UAA requests
* create new NewCloudControllerClient method and rename old one
* skipping test in windows due to timing issues
* update golang.org/x packages
* update ginkgo/gomega
* refactor ccv3 info test to reduce flakiness
* CF_DIAL_TIMEOUT works with api/cloudcontroller/cloud_controller_connection
* move API URL and SkipSSLValidation into TargetSettings struct
* add User-Agent header to CloudControllerClient
* return non CC errors in a generic wrapped error
* adds newline in run-task output
* clean up error handling
* add run-task integration tests
* rename requestOptions URI to URL
* remove v2 dependency in v3 commands
* refactor and backfill tests for run-task command
* add CCv3 run-task command
* remove build commands and link to Concourse task
* add file logging to new CF_TRACE
* add verbose logging to v2 commands
* add a request logger wrapper
* add ORG to usage message
* Merge branch 'hyenaspots-delete-space-o-redux'
* add comments about what we're skipping
* take care of some gometalinter warnings
* rename all the things
* change Make interface to take in http.Request in UAA Client
* change Make interface to take in http.Request
* clean up in api command
* exit early if target api and skip ssl configuration is the same
* reduced time taken to get all resources pointing to cli repo
* shortened wording
* reducing time taken by concourse to get the cli repo
* Revert "the integration package is not in the cli resource"
* remove the final-cli resource because it is not used
* re-add create-cats-config.bat to fix pipeline
* Bump version to 6.22.2
* Revert "Bump version to 6.23.0"
* Merge branch 'delete-space-o-redux' of git://github.com/hyenaspots/cli into hyenaspots-delete-space-o-redux
* Merge branch 'master' into delete-space-o-redux
* Update translation strings
* Update translation files
* Remove unnecessary/redundant translation strings
* Refactor delete-space to clean up after adding -o
* Add the results of bin/test i18n translation resource updates
* Add delete-space -o flag to V2 command list
* Re-add symlink fixture that was somehow deleted during rebase on master
* Fix bug needing space w/ same name in current org with -o flag
* Remove targeted org requirement from delete-space when -o is present
* Add -o flag (but still requires an org to be targeted)
* build fixed
* client password added
* Merge branch 'master' of https://github.com/cloudfoundry/cli into cf-oauth-client
* cf OAuth client added to config file

## 6.22.2
* Bump version to 6.22.2
* Revert "Bump version to 6.23.0"
* Bump version to 6.23.0
* adds task to debug windows installer in the release pipeline
* cats tasks get config from cats-config task
* Update windows and linux cats config setup
* Fix some issues with the cats-config job
* remove SKIP_SSO from CATs Linux
* reorder where CONFIG is set
* cats linux uses cats config task
* add task to generate cats config
* update contributing document
* update Slack incoming webhook
* Fix link to contributors' guide on README
* update translations
* add doppler endpoint to error message
* Update README.md
* fix the last link, hopefully...
* fix readme links, add how to debug cf curls for plugins
* change default timeout to 5 seconds for integration tests
* reduce number of files in repository root
* renamed docs badge
* added badges for docs and cmd ref guide
* fixed links
* added TOC to top of page
* use CF_DIAL_TIMEOUT to generate noaa timeout
* use http constants for http method names
* follow proxy if it is set in the environment
* wrap all request creation in NewRequest method
* better description for the path flag in unbind route service command
* replaced license badge url
* removed Appendix
* updated as per Chip's instructions
* increase cats-linux and cat-windows nodes to 10
* remove duplicate go get ginkgo
* set default username and password for cats-windows task
* inlined the stub to be compatible with echo
* r3 instances are not compatible with the bosh-lite AMI
* use the correct instance type
* change bosh-lite instance type to r3.2xlarge
* remove unused publish-rpm-timer resource
* use the doppler endpoint... for doppler
* use the token endpoint to talk to the UAA
* change create-buildpack position argument to int
* improve help description for bind-route-service --path option
* refactor v2actions.GetDomain
* on Windows the trace permissions come back as 0666
* change default to empty
* remove dora app from integration tests
* Update translations and unshadow error
* move GetRouteApplications to app actor, simplify route appending
* add integration tests for delete-orphaned-routes
* Extract GetSpaceRoutes and GetRouteApplications into actor methods
* Use ui.DisplayText for deleting route message
* separated out the ccv2.Route struct from the v2actions.Route struct
* Rename DeleteRouteByGUID to DeleteRoute
* Rename GetDomainByGUID to GetDomain
* Add / to path in test
* Update delete orphaned routes integration tests with patterns from unbind service
* Route paths already have a leading slash
* Remove query nil check
* godoc comments for actor methods
* corrected the godoc comment for CheckTarget
* backfilled tests in the actors and command packages
* adds queries for api methods
* backfills tests for the api layer
* additional test for String method on Route
* handle the error properly
* update the delete-orphaned-routes tests after rebase with master
* adds the actors and api layers for delete-orphaned-routes command
* Update delete orphaned routes actor methods to be consistent with other
* Add test for experimental warning
* Add delete orphaned routes command
* Add boolean user prompt function to UI
* reduce nodes to 3 :(
* reduce down to 3 :(
* backfill tests for cloud controller connection
* Delete Windows Diego CATs tasks
* switch to 5 nodes for DEA tests
* opt into garden-runc-release
* switch to using 6 nodes for CATs tests
* default dies-lite to diego backend
* switch everything over to use garden-runc
* Merge pull request #974 from iplay88keys/master
* skipping test
* Increase default timeout for Diego CATs
* Remove -r from CATs ginkgo command
* added hyperlink to New Issue page
* force empty lists for plugin app model
* changed the wrong timeout, should be default timeout instead
* increase the push timeout which is causing our linux CATs tests to fail
* rename get methods to follow conventions
* skips a cats app test that depends on tar and is breaking in our windows environment
* user provided services can now be unbound
* add godoc to API package
* add api call to get service instances from space endpoint
* use the correct path
* run windows unit test script from cli-ci instead of cli
* skip integration during units on windows
* skip integration tests in bin/test
* pull integration before running tests
* run integration tests from cli repo on Linux and Windows
* move integration test from cli-acceptance repo
* better error for unset org or space
* move target check to NewCloudControllerClient
* Reset timeout in callback
* Add localization resources for timeout err message
* Stop retrying the loggregator connection after 15 seconds
* add new line to display
* separate out the UAA client
* Skip config_test.go in CATs
* Use flag arguments to generate bosh-lite manifest
* Move api routes to internal and export APIRoutes
* generate bosh-lite cf manifest for routing-release
* refresh logic to the authentication wrapper
* rename routes to api_routes
* TailingLogs does not exit when noaa returns RetryError
* separate wrappers and connections into their own packages
* use specific error for an invalid auth token
* rename package api/cloudcontrollerv2 -> api/cloudcontroller/ccv2
* move error handling into a private wrapper
* add authorization to CC requests
* setup should always be a pointer reciever
* switch to string check as these error differ by platform
* refactor unbind-service command
* Use Windows environment variable substitution
* Translation placeholders for CF_DIAL_TIMEOUT help
* Split Windows GATs into separate tasks
* Readd CF_DIAL_TIMEOUT to 'cf help -a' output
* remove skipPackages ginkgo flag from pipeline and tasks
* Use acceptance BOSH public IP instead of private IP
* use the pipeline tasks from cli-ci resource instead of the cli resource
* Do not run v3 CATs on Windows
* update descriptions for restage and restart
* Remove https:// from API url in Linux CATS
* Use https in CATs config
* Remove https:// from api URL
* Merge pull request #966 from iplay88keys/master
* Use noaa token refresher and retry logic for logs
* Update go repo name
* added note to ask questions on CF Dev or slack
* corrected link to release notes page
* Handle app environment being nil in manifest
* Replace init() function with more idiomatic var _
* shows help for a subcommand when running cf with missing required arg
* swapped order of release notes and 32 bit releases
* updating resources maybe?
* friends don't let friends make packages called errors
* break config into separate files
* moved the ui package into the utils subdir
* path should be a string, not an int
* use correct filenames
* add windows installers back in
* Remove dependency on i18n for TranslatableError
* Fix 'Disbled' typo
* revised download section to include apt-get and yum
* added version badge, removed travis CI badge
* disable signing of RPM metadata for now
* fix directory
* use the correct directory
* fix errors
* windows installer creation still does not work
* use cli instead of cli-ci

## 6.22.1
* Bump version to 6.22.1
* remove depths
* move rpm publish into release pipeline
* fix indentation
* sign and create binaries and installers on windows
* reintroduce backwards compatibility flag to bind-route-service
* Add osslsigncode to ci container
* update homebrew-tap after release
* cli-ci should be independent from the rest of the pipeline
* read version prior to upload-releases
* Bump version to 6.22.0

## 6.22.0
* Bump version to 6.22.0
* runs the old api command unless EXPERIMENTAL is set to true
* adds 'Experimental' environment override to the config
* GPG key id is stored in a file, not passed directly
* follow symlinks
* add certificates to the debian publising process
* adding debian repo creation to release pipeline
* Add space after equals sign
* pull in cf-cli from previous step
* update names, trigger claw update when signing is complete
* add task for adding new version to claw env
* don't go chasing waterfalls
* upload to new-release-process for temporary testing
* because gpg is dumb and requires interaction
* try defining the macro directly
* add rpm config for signing
* force cleanup on exit, set the GNUPGHOME to working dir
* add tacker story delivery
* forgot to add the evil thing
* evil thing I need to do until vito/houdini/issues/6 is fixed
* improved wording
* clarifying these sections are for cli & plugin developers
* fix link
* Replace build script in compilation instructions with commands
* Add instructions for compiling the binary
* fix file location, don't run sign jobs in aggregate
* Show command help when given -h with command and extra args
* use lastpass to store credentials
* Remove duplicate invalid option message
* update translations
* adding base assets for release pipeline
* update english help text to remove inconsistencies
* adds environment variables section to the help command
* change PrivilegesRequired back to none from lowest
* add cli-ci add an input into cerate installers
* move installers and VERSION into ci directory
* enable installation of 32-bit windows cli as non-admin
* enable non-admins to install without admin password
* rpm use sha1 for checksum
* more godoc!
* name should align with external binaries
* archives created by concourse
* use the correct output directory
* no longer require gpg key
* no longer sign in our pipeline
* Add RPM repo publish task to pipeline
* remove signing from pipeline
* download the installer rpm
* Concourse task to generate and publish RPM repo metadata
* use elastic ip addresses for the bosh lites
* use relintdockerhubpushbot because it is more up-to-date
* export prior to changing directories
* run vagrant up in bosh-lite directory
* use the correct env variables directly in bosh-lite provisioning
* remove the correct value this time
* Revert "update to use a separate subnet"
* update to use a separate subnet
* have to set subnet, can set instance type via ENV
* try not changing the vagrant file, also destroy-on-error
* swap private ip for public one
* try automatically setting elastic ip
* Revert "pass along public ip address"
* pass along public ip address
* update debian repo location
* rename Debian keys to generic manager keys
* update cf cli release bucket
* update docker image to include createrepo
* sign rpm installers
* add options to cf plugins usage
* add options to cf org usage
* add options to cf space usage
* use a more concise error message when app name and manifest are not provided
* initial rewrite of 'api' command
* use the original error not the coerced one
* consolidate indentation for new help command
* workaround mktemp for old darwin
* use $TMPDIR instead of mktemp for osx
* Remove incidental mistranslations
* update error interface to adhere to standard error interface
* UI no longer merges map, config is written to disk on command exit
* verify that new phrases can be added to i18n files
* reduce ginkgo test run parallelism to ensure it does not exceed resource limit
* fix CF_API endpoint on gats scripts
* run new integration suite with GATs
* force posix style flags on windows
* ensure locks get released on all deploy-cf-release
* work around for the set-env
* fix test to check for usage errors on stderr
* display flag/argument errors to the screen
* patch go-flags to allow custom flags to read '-'
* use custom flag for memory and update go-flags
* use the correct arguments for UpdateBuildpackCommand
* update gats install path to code.cloudfoundry.org
* change padding on common help to 4 spaces
* org, not com.
* add travis test branch
* Update to code.cloudfoundry.com, and go 1.7
* add basic tooling for new i18n workflow
* update cli repo path to code.cloudfoundry.org
* Update lint debt
* fix spacing for create/delete-route commands
* properly use 2 spaces between common commands
* re-add line break between command categories for cf help -a
* fix math for print plugin commands
* adjust order of common plugin commands
* fixup i18n files
* refactor new UI so that it's injectable
* Revert "somehow these tests didn't get merged in"
* somehow these tests didn't get merged in
* make tcp route help port numbers consistent
* fix mismatching usage help for bind-running-security-group
* fix capitalization of examples on auth help
* add missing ds alias for delete-service
* fix wrong capitalization on delete-service help
* sort command help options
* cf help shows only common commands
* Remove FakeUI and added a NewTestUI()
* sort related commands
* add related commands to help text
* translate "SEE ALSO:"
* initial related commands structure
* somehow missed these in the last commit
* yeah, lets not keep debug output
* EndPoint -> Endpoint, UaaEndpoint -> UAAEndpoint
* add translations and fixed up some resources
* fix help headers translation
* fix language translation order
* detect binary name from invocation
* internationalization is back
* displays installed plugins help by plugin and sorts commands alphabetically
* use Alphabetic
* Update test to reflect error returned from go-flags
* Fix colors on Windows
* switch to code.cloudfoundry.org
* don't display plugins after every help category
* support '-v' properly
* help displays plugin specific help
* add installed plugin commands to master help text
* read plugin config
* support the -v/--version flag
* gofileutils/fileutils uses code.cloudfoundry.org
* read config prior to running commands
* added full page help
* add actor interface
* support additional error cases in help
* add usage to all commands
* show help with aliases and incorrect usage errors
* WIP Backfilled all commands
* use standard golang naming in commands directory
* handle the horribleness that is PanicQuietly
* move error handling to utils
* add all the commands
* introduce go-flags
* github.com/cloudfoundry/cli -> code.cloudfoundry.org/cli
* update untranslated files
* fix translation issues
* Merge remote-tracking branch 'origin/pr/940'
* winstallers: fix include directive for x86 installer
* winstallers: fix source path for cf.exe
* winstallers: fix broken include directive
* winstaller: add cf.exe to Path for non-admins
* add default install directory
* allow user to change install directory
* change default install directory for non-admin users
* build windows installer does not require privileged user
* Merge remote-tracking branch 'origin/pr/926'
* Merge pull request #936 from hyenaspots/contrib_doc
* add shipt job to gate resources before release
* preserve timestamp when copying cf binary on release
* Revert "skip copy file to preserve file modification timestamp"
* Revert "fix minor issue"
* homebrew tap only commit if sha changes
* fix minor issue
* skip copy file to preserve file modification timestamp
* Update i18n resources
* make checksume constant when compressing the same file
* added license shield
* Improve contributor's guide
* fix bad yaml
* add sha256 signing to gpg
* fix gpg_key printing on publish_debian task
* Change import path of gpg private key
* match os_arch on .deb filename
* extract tarball with .deb before processing
* edit path again for debian version
* fix path for publish debian
* configure more debian ci
* publish debian repo in publish final release job
* add directory before untarring installers
* Non-interactive ruby installation on ci
* Use private repo to import gpg key
* add more spacing for concourse
* add a space
* add debian publishing to ci
* Prevent usage messages form being printed twice
* fixes an error with the way the cli resolves symlinks on windows
* improved wording
* moved edge binaries to end of download section
* added plug for/link to community plugins site
* added filing feature requests
* removed details that are included in the issue template already
* added items from main readme
* mentioned that PRs with command/flag changes should be discussed in an issue first
* Merge pull request #931 from cloudfoundry/shalako-patch-1
* Fix a few problems that arise with go 1.7
* output stderr for plugin commands
* no more shadows
* fixup bad linebreak on i18n file
* Allow apps with 0 instances to be started
* trace: Redact all passwords and tokens in JSON
* remove more panics
* command requirements use errors rather than panics
* more fix
* fix bug for windows appfiles test
* Enable compression when creating zip files
* Ensure cf curl output is properly passed back to plugins

## 6.21.1
* Bump version to 6.21.1
* update diego resources to come from cloudfoundry, not incubator
* Add CONTRIBUTING guide and cleanup README
* configure dies' bosh-lite doppler to port 443
* create bosh lite cf manifest can take an extra stub
* chmod windows binaries before combining
* shortened line with links to help
* added slack url
* changed bugs to issues
* simplified login instructions by using login's -a option
* merged Getting Help section into top
* added known issue with cygwin
* added links to docs, command ref guide
* add i18n resources
* update error message
* move binary to the right output
* fix pipeline config
* adding task to sign windows binary
* modify error message for files command
* create and validate local builpack zip file before api calls
* print standard log out without color
* curl command now requires api endpoint set
* Expose CF_DIAL_TIMEOUT to help text
* fix error when uploading buildpack using url
* no more invisible log text in windows terminals
* Remove dummy pipeline from master. Move to remote.
* Fix up some inputs and outputs for dummy 6.19.0
* create tmp pipeline that makes 6.19.0 win64 installer
* package correct binary in windows installers

## 6.21.0
* Bump version to 6.21.0
* add validation for --no-hostname
* Revert "move validation after CLI parameter merging"
* move validation after CLI parameter merging
* support -d private-domain with new manifest routes
* fixing translation
* add -d to replace shared domain
* removed unnecessary periods from translation files
* Ensure cf api default url scheme defaults to https
* update i18n
* Change routes to map slice
* hyphenate to indicate single value input in German
* Fix translation regression in Chinese Simplified
* create_app_manifest will print routes instead of hosts and domains
* Skip route lookup when tcp route has a random port
* add route-path flag to app manifest push
* add usage of random route flag to push help text
* enable cf push --random-route for app manifest
* remove -b and --build
* Check hostname for private routes on app manifest push
* delete old pipeline
* use team email instead of Ted's
* does not set host when pushing manifest with tcp routes and -n flag
* no pipes please
* specify running instance
* forgot to update path in windows tasks again
* forgot to update path in windows task
* enable -n flag to change manifest route
* forgot jobs key in pipeline.yml
* reorganize ci directory
* Remove unused translations and clean untranslated list
* Merge branch 'simonleung8-master'
* Update language files
* update cli-plugin-repo
* update kr/pty
* update gorilla/websocket
* udpate go-querystring
* update sonde-go/events and other dependencies
* update bmizerany/pat
* update blang/semver
* update cloudfoundry/gofileutils
* update go-i18n [#126537363]
* Update docker and other dependencies
* switching branch of github.com/tedsuo/rata
* updating github.com/nu7hatch/gouuid
* updating vendor/golang.org/x/sys/unix
* updating golang.org/x/net/websocket
* update golang.org/x/crypto/ssh
* use 'Consistently()' with 'ShouldNot()'
* update yaml package
* update loggregatorlib/logmessage
* routes and no-hostname should not co-exist in the app manifest
* update lager
* remove dependency on diego-ssh/cf-plugin/terminal
* update ginkgo and gomega
* flags -> cf/flags
* json -> utils/json
* Increase timeout on cmd suite
* glob -> utils/glob
* generic -> utils/generic
* downloader/ -> utils/downloader/
* commandsloader -> cf/commandsloader
* more hidden cf directories
* new bosh-lite pipeline
* validate manifest before merging with flag context
* validation for http/tcp routes with port/path
* make cats-diego blocking
* move cmd from TL to cf, and spellcheck into utils
* update the tests to reflect the changes in error messages
* stack name displays correctly on cf push
* clearer error messages for manifest validation
* move words -> utils/words
* Revert "Added a self signed cert to the acceptance cf deployment"
* we do in fact use that input for the tasks
* for some reason the file is now mysetup.exe
* support application manifest with routes field that has ports
* Added a self signed cert to the acceptance cf deployment
* Configure the acceptance pipeline to associate an
* Configure the acceptance bosh-lite deployment to
* Adds task to associate elastic ip with an aws instance
* Revert "switching to 7z as well, sourceforge still broke"
* Revert "swap GnuWin32 with 7zip, cuz sourceforge is broke"
* go back top using tar and zip from gnuwin32
* Change domain to system_domain
* update windows create-installer to use 7zip
* remove unneeded godeps, add some debugging
* Break up the deploy-bosh-lite job for faster freedback cycle
* app manifest accepts routes attribute with paths
* MapManifestRoute test wasn't testing anything
* Deploy the bosh-lite to its own subnet
* app manifest parsing should fail with empty domains and routes set
* capitalize error messages
* Bubbles up the correct error msg from the API
* add close connection to http header
* indent ensure & on_failure to act on task level
* Add functionality to create and bind routes from a manifest.
* use max api version variables in commands
* use min api version variables in commands
* parallelize all the things again
* main -> cmd, run cli as a library
* refactor route actor to be injected into push
* use default foreground color instead of white
* Use Normal White instead of Bright White
* Add slack-alert resource
* Use Eventually().Should(Exit()) instead of Expect(Exit()).To(Equal())
* update-brew-formula task calculates the correct sha256 hash
* fix update brew formula script

## 6.20.0
* Bump version to 6.20.0
* Use dial timeout for UAA and Routing APIs also
* properly handle '.*' in .cfignore
* Make bind-security-group help text clearer
* Set the timeout in the correct place
* return cli execution panics as errors for plugins
* Make polling throttle configurable
* Prevent negative waitgroup counter with the noaa logs repo
* stop ignoring plugin dir during linting
* Fix path on windows worker for ginkgo.
* backing out of these changes temporarily
* Update concourse to 1.3.1
* true=keep going, and remove unnecessary newlines
* validate manifest yaml on push
* roll back remove panic from app command
* Remove some panics from requirements checks
* 'routes' can be parsed by manifest
* clarify copy source help message text
* bind security groups to all spaces in org
* Remove panics
* Remove RunCLICommand from push_test.go
* Replace old spacefakes with current fakes
* add route path to unbind-route-service
* Add --path to bind-route-service
* Fix errors created in 6a5be90
* Remove testhelpers/terminal from cf/net
* Explain where the Writer is assigned
* introduce constructor to panicprinter and remove testhelpers
* Always output panic error exceptions
* use common cats linux task for diego tests
* remove time from testhelpers
* improve formatting of log messages
* run CATs services along side the rest of CATs
* missed the last comma
* add a way to ignore the default tests
* Remove testhelpers/cloudcontrollergateway
* Explicit returns
* remove testhelpers/assert
* Remove race from ./plugin_examples
* Fix race condition in ./plugin
* Remove the FakeReqFactory.
* Use counterfeiter fakes in cf/commands/user
* Use counterfeiter fakes in cf/commands/space
* switch service key to counterfeiter
* Use counterfeiter fakes in cf/commands/servicebroker
* switch serviceauthtoken to use counterfeiter
* switch service access to use counterfeiter
* Use counterfeiter fakes in cf/commands/service
* Use counterfeiter fake in cf/commands
* Fix data races in RPC/Plugin code
* switch security group to use counterfeiter
* switch security group to use counterfeiter
* switch routes to use counterfeiter
* switch quota to use counterfeiter
* Adds a concourse task to update the homebrew tap
* Actually fix data races in cf/commands/application
* Update to new version of go-ccapi with LICENSE
* remove unnecessary test, fix other one
* switch plugin and pluginrepo to use counterfeiter
* switch organization package to use Counterfeiter
* switch feature flag to use counterfeiter
* Fix data races in cf/api/logs
* Fix data races in cf/commands/application tests
* Add locks to FakeUI, ProgressReader
* switch environment variable group to use counterfeiter fakes
* switch domains package to use counterfeiter

## 6.19.0
* Bump version to 6.19.0
* Use proper coloring
* move -race into units only, also randomizeAllSpecs
* update commands to use counterfeiter fakes
* update buildpacks to use counterfeiter fakes
* Upgrade noaa library
* Suggest commands for mispelled plugin commands
* set ssl-validation to true
* moved around variables, use http for CATs tests
* set user/password for cats tests
* use release-integration's run cats test task
* fix disable-service-access, stop displaying additional help text
* Race detected tests by default
* Upgrade loggregator_consumer
* Use master for routing-release
* Add command auto-suggester
* Rename acceptance pool in dies-lite
* Add dies-lite pipeline
* Remove vetshadow lintdebt
* Make bosh-lites-pipeline accurate
* Remove lintdebt from plugin_examples
* bosh-lites-dea pool is actually on master
* Switch dea pool to cli-pools
* Switch diego cats back to 'bosh-lite-lock'
* Switch bosh-lites pipeline to use cli-pools
* Remove named return values
* Properly unit test set-space-quota
* Use counterfeiter fakes spacequota package
* Use counterfeiter fake in application package
* Adds missing error handling
* Add errcheck slow linter
* Add space requirement test and remove imports
* Refactor organization command tests to not use runCLICommand
* Refactor space tests to not use runCLICommand
* Configure the lint deadlines
* Add slow linter: unconvert
* Add deadcode linter
* Add vetshadow
* Add gocyclo linter
* Add gofmt linter
* Add vet linter
* Add NewNumberArguments to Factory
* Introduce NumberArguments requirement
* Remove two more Execute() panics
* Remove one more Execute() panic
* Provision dies lite
* Fix UnlimitedDisplay constant
* Does not display route ports when not provided by API
* Does not display route ports if API does not provide it
* Display route ports for space quota in cf space
* Shift the indentation
* Show reserved route ports in org quota in cf org
* Change indentation of test
* Skip if exists
* Update routing release acceptance test name
* Adding in goconst linter
* Switch to gometalinter
* Make casing consistent
* Use all these magic flags to build static binaries
* Handle language codes zh-TW and zh-HK
* Panic -> DoesPanic
* Add color for space quota
* Eliminate stutter debt
* Use CC error for invalid reserved-route-ports limit
* Remove last of the debt
* Remove more debt
* Remove more debt
* Remove more lintdebt
* Remove more debt
* Remove lint debt
* Remove more debt
* Remove some debt
* Remove unused test helpers
* Add codeclimate to travis
* Enable codeclimate analysis
* Manually revert ad587e7e91121c307e7d08fb935a42648b684c59
* Revert "Remove inline-relations-depth from cf organizations"
* Use git-hooks style for hooks
* Don't release lock if stopped trying to connect
* Improve tip for set-env command
* Add route ports to space quotas command
* Refactor space quotas to use FormattedServicesLimit
* Add reserved route ports limit to space quota
* Move services limit formatting to model
* Initialize structs better
* Release wait group after printing to not risk interweaving output.
* Remove named return
* Remove shadowed variable
* Command.Execute() returns an error
* Update total_reserved_route_ports on quotas
* Update ginkgo and gomega
* Remove more potentially offensive words.
* go format
* Remove potentially inappropriate words
* Configure max TCP routes on create space quota
* Remove RunCliCommand from create-space-quota
* Remove unused dependency

## 6.18.1
* Bump version to 6.18.1
* Ignore errors from printing/syscalls in cf/terminal package
* Explicitly ignore errors in cf/ssh package
* Handle errors in cf/net package
* Handle errors and remove named return values
* Explicitly ignore failed Flush
* Remove named return values
* Explicitly handle errors
* Display 'unlimited' instead of -1 for org quotas
* Remove multi-byte colons
* Use actor when actor_name is empty in `cf events`
* Remove RunCliCommand in application/events_test.go
* Parse actor from events fields
* Translate APP_NAME
* Display flags in marketplace help text
* Merge branch 'master' of https://github.com/simonleung8/cli into simonleung8-master
* Display bound apps for user provided services
* Do not hardcode ':' for some languages to be translated correctly
* Add back translations that were deleted
* 64bit windows installer now installs in correct directory
* Running cf service service-name now displays 'Bound apps: '
* Remove RunCliCommand from service_test
* Remove naked returns
* Merge branch 'master' of https://github.com/simonleung8/cli into wip/#119616001-ibmtranslations
* Remove shadow
* Skip chmod for windows because it is not supported
* I TAKE FULL RESPONSIBILITY FOR MY ACTIONS
* Order these tests to prevent test conflits
* Fix gometalinter issues in cf/commands
* Handle writer.Close errors in cf/api
* Remove named return from performMultiPartUpload
* Fix gometalinter issues in cf/actors
* Update i18n languages
* List service instance bindings by instance guid
* Remove named returns and use ghttp in tests
* Simplify lookup for all of service's plans
* Formatting
* Simplify service plan lookup
* Remove inline-relations-depth from cf organizations
* BackwardsCompatibility flags do not require args
* Missed a spot
* Use cli-ci to run tests in windows gats
* Randomizes the GATS test suites
* Diego tests now run as part of the pipeline
* Remove dependence on org returned from space api calls
* Remove Comments
* Fix Path
* Backwards compatible `cf bind-route-service -f`
* Request spaces to be ordered alphabeticaly
* Rename cc_fake to old_cc_fake
* Remove inline-relations-depth query from cf spaces
* ignore the actual work for now and lets see if the basics work
* Remove duplicate ginkgo install
* Switch diego tests to use cli-pool resource
* bosh-lite pipeline now uses bosh-deployment resource
* Use the cli-pools resource for realz
* Revert "Use the cli-pools resource"
* Use the cli-pools resource
* Use docker image with bosh cli
* Use only what we need from the generate-deployment-manifest
* Display only plugin filename upon install of remote plugin
* Handle extra spaces in download file headers
* Deploy diego task uses cli-pools
* Removes diego deploy dependency on cf-release repo
* Generate stub file after git clean
* Make bosh commands non-interactive
* Add sleep since the network might not be up
* bish bosh borsh
* Fix heredoc in provision-cf-lite task
* Changes the provision task to generate a bosh-lite manifest
* Get organizations alphabetically
* Refactor org tests to no longer use old handlers
* Bosh-lite 4 now uses new diego pool
* Add default to diego bosh-lite script
* Remove -f flag from binding route service
* Merge pull request #839 from fujitsu-cf/update-ja_jp-files
* Japanized Bind cancelled, Creating route, Creating service broker and some incorrect usage mesages.
* Merge pull request #838 from MarkKropf/master
* Update bump-version script
* reduce the promiscuity of our random words

## 6.18.0
* Bump version to 6.18.0
* Aggregate gats-linux + gats-windows into gats
* Run GATS in parallel
* Default windows to use colors
* Add cats-windows-diego job
* once, only once
* Run new cats-linux-diego job
* stop lying about the help
* Readme documents plugins not being passed the -p/-h flags
* Parallelize gats and cats
* Remove default from -a help in `cf update-space-quota`
* Shorten header names of `cf space-quotas`
* Hide ReservedRoutePorts for quota if not in API call
* run with two nodes, so half the time?
* Shorten header names for `cf quota`
* Make plugin uninstall description more prominent
* Ratcheting is somehow failing on CI
* Fix some lintdebt
* Update lintdebt
* Check for gofmt in pre-commit hook
* Add ratcheting to our pre-commit hook
* Update ratchet command output
* Automate ratcheting during bin/test
* Add bootstrap script and generic pre-commit hook
* Update lint ratchet
* Remove defaults for flags on update quota
* Add route ports column to quotas table
* Oops, it was complaining about not being a pointer
* Fix linter error
* Fix flaky test
* Eliminate RunCliCommand in help_test
* Limit OS specific Writers to main
* Inject OS specific writer
* Support reserved route ports flag on cf quota
* small things are better
* Add minor support for a team custom linter
* updated link to http://docs.cloudfoundry.org/cf-cli/use-cli-plugins.html
* add reserved route ports line to cf quota table
* Implement reserved route ports flag on route quota
* Add requirement for the new flag.
* Move flag and function tests to when user is logged in
* Add reserved-route-ports flag help messages to update quota
* export UpdateQuota to backfill help message tests
* Add required arg for --reserved-route-ports help message
* Update interface to use new type
* Convert role const to Role type
* Fill in --reserved-route-ports functionality for create-quota
* Add minimum API requirement for --reserved-route-ports
* Backfill requirements tests to CreateQuota
* Add flag & help text for --reserved-route-ports
* Backfill command help text tests
* adding verbose version for windows too
* Adding a verbose-debugging pipeline for gats
* Use official routing release instead of building ourselves
* Force all jobs to occur on non-check-worker.
* Return to Loggregator team's NOAA library
* Simplify color logic
* Fix display of tcp routes in delete-orphaned-routes
* Use https to fetch from final-cli repo instead
* Fetch tags for publish-final-release
* Login first, it helps!

## 6.17.1
* Bump version to 6.17.1
* Set the deployment prior to running smoketest
* Run smoke tests after deploy-boshlite
* Backfill CHANGELOG.md for v6.17.0
* generate-changelog writes to CHANGELOG.md
* suggestions made by 'gofmt -s'
* Use InnoSetup to sign the generated windows exe installer
* All the nodes (except only 2) and for CATS
* No nodes for gats linux
* Pump up the jam, pump it up/ While your feet are stompin'
* Add cflinuxfs2-rootfs-release to deploy-diego
* Can you C what I did here? :grin:
* Formatting
* Fix capturing terminal output on windows
* WHAT AMI!? I CAN'T EVEN?!
* Run tests in parallel
* Up the size of bosh lite
* Fix err variable shadowing govet error
* Staticly link Linux binaries
* rename edge installer to match version installers
* Taste the rainbow (bring colors to windows)
* Check for existence of CF_HOME
* Update README
* Revert "Windows x64 installer puts things in correct directory"
* Remove data races from NOAA log repo
* Windows x64 installer puts things in correct directory
* added request to check the latest version
* Merge pull request #821 from andreas-kupries/ticket-ref-117404629
* Adding back in noaa library
* Removing noaa, so hopefully submodule doesn't show
* Using gvt to Recreate vendor directory
* Fix collateral damage
* Important to flush (message queues, that is)
* Remove tests from noaa package
* Merge noaa to master
* Merge master into wip/104053574-switch-to-noaa
* Change error message to specify command instead of RoutingAPIRequirement
* Update noaa to challiwill fork
* Ui -> UI
* Api -> API
* Cli -> CLI
* Do not break plugins
* Url -> URL
* Uri -> URI
* Guid -> GUID
* Update acronyms
* Only display insecure warning for insecure api's
* Ref ticket 117404629. Reviewed transforms done on a cell value. Had to trim value before transformation in all places (*). Moved application of the transform into printCellValue, to keep things nearer to each other and make comparison of printing and width calculation easier.
* Copy old translations and update them to 'https'
* Use new noaa consumer library
* Update noaa library and some others
* Linter won't complain
* Make returned vars explicit
* Add back properties in stub
* Fix routing api property
* Remove RunCliCommand
* Enforce https connections to plugins.cloudfoundry.org
* Use https:// in plugin help text
* Put property-override where it won't get deleated
* Make consistent with deploy-routing-release
* Set appropriate tabbing
* Make formatting correct
* Use https for plugin-repos
* Add LITE_HOSTNAME back
* Add property-override stub back
* for concourse to pick up previous commit
* Update provision cf script and remove unnecessary envs
* Update final-cli resource path
* Noaa Log repository added
* Changed LogsRepository interface
* Moved FakeLoggregatorConsumer
* Introduce generic interface for LogsRepository
* Adding noaa library to deps

## 6.17.0
* Bump version to 6.17.0
* Update TcpRoutingMinAPIVersion to 2.53.0
* Add reconfigure script for cli, bosh-lites, concourse-redeploy
* Remove outdated update to credentials.yml
* Remove extraneous MinAPIVersionRequirement from routergroups
* Hide "app ports" column
* Add minimum version requirement to router-groups
* Trim columns when figuring out padding length
* Update french translations
* Move router-groups into DOMAINS section of cf help
* Fix RoutingApiRequirementTest
* Update RoutingApiRequirement
* Create Requirements slice
* Remove app-ports from tests
* Hide app-ports features
* Add hidden setting to flags
* Merge branch 'refresh_config_after_requirement_failure'
* Refresh API version before failing due to version
* Output go version in build jobs
* Don't install go vet, build in after 1.5
* Map routes will take in a random-port flag when passed
* Map routes will take in a port flag when passed
* Merge branch 'remove_underscores_from_packages'
* github.com/cloudfoundry/cli/cf/actors/plugin_installer
* github.com/cloudfoundry/cli/cf/actors/plugin_repo
* github.com/cloudfoundry/cli/cf/actors/service_builder
* github.com/cloudfoundry/cli/cf/api/app_events
* github.com/cloudfoundry/cli/cf/api/app_instances
* github.com/cloudfoundry/cli/cf/api/application_bits
* github.com/cloudfoundry/cli/cf/api/copy_application_source
* github.com/cloudfoundry/cli/cf/api/environment_variable_groups
* github.com/cloudfoundry/cli/cf/api/feature_flags
* github.com/cloudfoundry/cli/cf/api/security_groups
* github.com/cloudfoundry/cli/cf/api/space_quotas
* app_files -> appfiles
* github.com/cloudfoundry/cli/cf/command_registry
* github.com/cloudfoundry/cli/cf/configuration/core_config
* github.com/cloudfoundry/cli/cf/configuration/config_helpers
* github.com/cloudfoundry/cli/cf/commands/plugin_repo
* github.com/cloudfoundry/cli/utils/testhelpers/rpc_server
* github.com/cloudfoundry/cli/cf/configuration/plugin_config
* github.com/cloudfoundry/cli/utils/testhelpers/plugin_builder
* Update bin/test
* github.com/cloudfoundry/cli/utils/testhelpers/cloud_controller_gateway
* github.com/cloudfoundry/cli/commands_loader
* github.com/cloudfoundry/cli/cf/ui_helpers
* github.com/cloudfoundry/cli/cf/panic_printer
* github.com/cloudfoundry/cli/cf/command_registry/fake_command
* github.com/cloudfoundry/cli/plugin/rpc/fake_command
* Remove parallel from bin/test for speedup
* Rename all fakes
* github.com/cloudfoundry/cli/cf/actors/plan_builder
* github.com/cloudfoundry/cli/cf/actors/plan_builder/plan_builderfakes
* github.com/cloudfoundry/cli/cf/actors/broker_builder
* broker_builderfakes
* Use port flag when unmap-route
* Deny specifying both port and hostname/path for route unmap
* Specify min API version when unmapping route
* Update usage/help text for unmap-route with TCP routes
* Regenerate binary only once per sweet suite
* Show path/port in route summary for `cf app`
* Create new CreateRandomTCPRoute actor function
* Return name first in app manifest
* plugin rpc package uses new counterfeiter fakes
* requirements package uses new counterfeiter fakes
* trace package uses new counterfeiter fakes
* words generator package uses new counterfeiter fakes
* testhelpers rpc_server package uses new counterfeiter fakes
* v3 repository package uses new counterfeiter fakes
* ssh package uses new counterfeiter fakes
* utils package uses new counterfeiter fakes
* Net package uses new counterfeiter fakes
* manifest package uses new counterfeiter fakes
* plugin_config package uses new counterfeiter fakes
* Configuration package uses new counterfeiter fakes
* commands and route package uses new counterfeiter fakes
* applications package uses new counterfeiter fakes
* CommandRegistry package uses new couterfeiter fakes
* Appfiles package uses new couterfeiter fakes
* API package uses new couterfeiter fakes
* SecurityGroups/... uses latest counterfeiter fakes
* Replace old counterfeiter packages from most of api
* Remove the commented blue.
* Update translations for `cf push`
* Update Domains to singular router_group_type
* Forgot to update translation files
* Make this test pass in a tty session
* Shouldn't need to manually install go vet anymore
* Actors now use latest counterfeiter fakes.
* Update path option help text
* Colocate Printer interface with UI
* Do not leak *[]string abstraction into TeePrinter
* Move interfaces/fakes to call site
* Extract UITable from Table
* Allow TCP random port when cf push with --random-route
* Merge pull request #811 from andreas-kupries/hcf-466-cli-extended-table
* Update the translation files
* Show port when create/deleting routes
* Use the passed in port when finding a route to delete
* Gate port deletion behind min CCAPI version
* Fail if we provide a port and hostname/path
* Update delete-route help text for TCP routes
* [HCF-466] Extended cf/terminal/table: Multi-line cells. Suppress header row. Custom per-column transformations. Extended testsuite.
* Do not wrap env vars in quotes
* Use response from /v2/shared_domains for route types
* Update fakes.
* Do not use global rpc.DefaultServer in testhelpers
* No need for a makefile
* Remove unused and outdated scripts/docs
* Fix LICENSE
* Update bosh-lite-locks to cli-private
* cli-ci -> cli-private rename concourse-redeploy
* cli-ci -> cli-private for bosh-lite pipeline
* Rename cli-ci repo to cli-private
* Routing team renamed release_candidate to release-candidate
* Use CATS's own test script
* Do not default to diego backend
* Set deploy-routing-release to run after diego
* Remove workarounds in deploy-routing
* Test routing release after deploy
* Use tar correctly
* Fix tgz & output directory
* Correct input directory for combine-binaries
* Don't cross compile OSX binary
* Use the correct names
* OSX binaries are built on OSX machine
* Pipelines no longer cause deprecation warnings
* Display unlimited instead of -1
* Show app instance limit on spaces, space-quota(s)
* Get space quota returns app instance limit
* Add `-a` flag for {create,update}-space-quota
* Quota update should not reset app instance limit
* Renamed AppInstanceLimitMinimumApiVersion -> OrgAppInstanceLimitMinimumApiVersion
* Use the unlimited instances constant
* Fix "invalid input" error when calling curl -d with an empty string
* should have run the tests
* add paramater translations in create-route
* display port on `cf create-route --random-port`
* Return app-ports in manifest only when they exist
* Add app instances usage in {create,update}-quota
* Update translations for update-quota
* Split usage translations for create-quota
* Display app instance limit in `cf org`
* Display app instance limit in `cf quota`
* Return app instance limit quota in `cf quotas`
* Trigger deploy diego after deploy cf
* User can set total number of app instances on quota
* use idiomatic checking of flags
* Rename misnamed create/delete space quota files
* cf curl -d with empty body makes a POST request
* add concourse-redeploy pipeline
* Give create-service-broker command alias "csb"
* Remove other No Arguments with usage
* include app-ports in generated manifest
* Fix regression where CF_TRACE/config.trace prevents terminal output
* We are better than this
* Introduce UsageRequirement for "No argument" usages
* Remove 'Incorrect Usage. No argument required.'
* Reverse polarity of UsageRequirement predicate
* Remove 'Incorrect Usage.\n\n'
* Remove 'Incorrect Usage\n\n'
* Extract a UsageRequirement
* Update interface of Requirements() in README
* Remove error from Command.Requirements interface
* Remove naked returns from Command.Requirements()
* Extract minimum API version from push with --route-path
* Deliver stories with concourse
* Eliminate double printing FAILED in verbose modes
* Fix test messages
* Remove reqFactory dependency on terminal.Ui
* Change Requirement.Execute() interface to return error
* user should be able to bind/unbind all routes
* Can update app-ports from manifest on existing app
* Use newest cats-windows.bat file
* Old cf-acceptance-tests uses Godeps/_workspace
* Use cf-release-repo as our GOPATH
* Only run cf-acceptance-tests bundled with cf-release
* Attempt to fix go vet issue
* Merge branch 'master' into wip/114265763-verbose
* remove unnecessary usage text
* Extract trace.Printer dependency to main
* make bosh-lites-pipeline more like a pipeline
* Run provision script from context with vagrantfile
* Don't fail bosh-lite deploy if there are no instances
* handle -h wherever it may appear
* Boyscout rule for query params package
* Move generate_port from req body to query params
* Update godeps to v57 && go1.6
* move trace logger a couple levels up out of ui
* push logger to top level of AuthenticationRepository
* push logger to top level of http client
* Add global -v flag
* Fix typo in Dependency
* Replaced Ω with Expect in main_test
* Update README to reflect changes to CommandMetadata
* Update Usage for the create-route command
* Fix test that generates manifest files
* Rename CommandMetadata.Example to Examples
* Extract examples from create-service-key command
* Extract examples for update-service command
* Extract examples for create-user-provided-service
* Add example text for create-service
* Extract examples from bind-route-service
* Update bind-service examples
* Merge branch 'master' into wip/114821419-extract-examples
* Add deploy-routing to bosh-lites pipeline
* Use existing CF deployment when setting diego/routing on
* Cannot use spiff for sslCert/Key in UAA
* Create UAA SSL workaround for routing
* Create comments to delete override cruft
* Fix worker-overrides filename
* Attempt to reduce compilation workers to 2
* Override domain again in cf manifest
* Fix EOF location
* Deploy routing release
* Enable diego release deploy
* Run non-interactive commands
* Enable diego_docker feature flag
* Also include cf-release for diego scripts
* Push into diego-repo not release
* Fix
* Skip uploads if already exists
* Deploy diego with manifest
* Add authentication to bosh
* Add task to deploy diego onto a bosh-lite
* remove unused resource
* extract bosh-lite jobs to separate pipeline
* Push create-app-manifest io to cmd.Execute
* Refactor generic.Map code
* Fix error on curl -d with empty string
* Update gitignore for VS Code
* Clean up AppParams merging
* Clean up cloud controller error codes
* Update windows ui to respect new interface
* Fix gats-linux
* Don't interpret '%' in ui output as printing verb
* Remove Godeps/_workspace from GOPATH
* Destroy bosh-lites before rebuilding
* Forgot the s3 in the command
* Update publish archive tasks to use awscli
* Remove the check to see if signed properly
* Sign the OSX installer in create-installers
* Remove all references to CodeGangsta
* Extract min version in all commands that use IsMinVersion
* go fmt language resources
* Use different test for changed i18n files
* Fix copy-source -h help text
* Look up space quota in specified org during `cf create-space`
* Run generate-language-resources less often
* Fix go fmt on push
* Hide --app-ports in cf push
* Speed up start command tests
* Fix usage text for cf push command (add --app-ports)
* Remove invalid default quota comment from create-space
* Update update-user-provided-service
* Improved routes test
* Display appropriate error when TCP route cannot be created
* Update delete-service-key
* Update unbind-route-service
* Update service-keys
* Update service-key
* Update unmap-route
* Update map-route
* Update delete-route
* Update create-route
* Update check-route
* Update remove-plugin-repo
* Update add-plugin-repo
* Update commands examples
* Update curl command
* don't empty tmp dir
* Update auth, login, repo-plugins examples
* Extract EXAMPLE from Usage
* Change Usage to a slice of usage strings
* Speed up main test [#114808063]
* Extracted TcpRoutingMinimumApiVersion from create-route
* Fix inputs on pipeline
* Empty windows tmp dir after successful installer
* Update USAGE for create-route
* Added `version` command
* Turned cf.Name into a variable
* Add 'app ports' to app output
* Add 'app ports' column to apps table
* Remove --build from help
* push supports --app-ports flag
* Rewrite ApplicationRepository.Create tests
* Update unbind-route-service description
* Update bind-route-service description
* Merge branch '6.16.1'
* create-route accepts --random-port flag
* Third time's the charm
* Try again with fixing the paths/include paths
* Separate resources for cli/ci and cli
* Add 'port' and 'type' to routes output
* Add 'type' to domains
* Fix deploy-boshlite tasks
* Remove unused installers/windows files
* Revive windows-installer.iss
* Fix RouteCreator interface
* create-route supports TCP routes via --port
* RouteRepository.CreateInSpace supports port field
* routes.CreateInSpace doesn't send empty path
* Update cats/gats bat files
* Improve performance of `cf logs` command
* Add ci/pipeline and provision-lite tasks
* Update Concourse tasks to use correct path
* Remove old CI files; replace with new ones
* Revert "Hide --router-group option to create-shared-domain"
* Revert "Hide router-groups command"

## 6.16.1
* Bump version to 6.16.1
* ssh-code returns an SSH code instead of auth token

## 6.16.0
* Bump version to 6.16.0
* Remove 'type' column from domains output
* Hide router-groups command
* Hide --router-group option to create-shared-domain
* Fix typo in pt-BR translation
* Replaced confusing url in `cf install-plugin` help
* Fix fr translation with extra space at end
* Hide .cf directory on Windows
* bin/generate-language-resources
* create-app-manifest stores stack
* Update travis to run on go1.6
* Don't translate role names
* Update translation files post-PR merge
* Move Dockerfile & Makefile to cli-ci repository
* Promote cliFlags package to flags
* Run goimports
* Remove unused code
* Sort command flags
* bin/generate-language-resources
* Merge branch 'update-ja-jp-messages-route-services' of https://github.com/fujitsu-cf/cli into fujitsu-cf-update-ja-jp-messages-route-services
* cf-Apps -> CF-Apps
* Wrap 'CLEAR', 'port', and 'none' in translations
* translate the messages to Japanese.
* Don't set GO15VENDOREXPERIMENT=1
* Simplify i18n/locale
* Fixed incorrect usage of e.g. in delete-orphaned-routes
* Add translation to invalid locale error message
* Fail when config is given an invalid locale
* Replace cf with CF when referring to Cloud Foundry
* Remove 'app' local var
* Add disk quota to create-app-manifest manifest
* Remove named return args from create_app_manifest
* Surround option values with single quotes
* Fix `list-plugin-repos` typo in help
* Added GitHub issue template
* Don't send empty path when searching routes
* Start command doesn't use appRepo.Read
* Remove unused arg from setup func
* Update French translation for plugins cmd description
* Switch non-English key back to English
* bin/generate-language-resources
* Update ja-jp untranslated json
* Merge branch 'update-ja_jp-files' of https://github.com/fujitsu-cf/cli into fujitsu-cf-update-ja_jp-files
* Update ja-JP files
* Update Dockerfile for latest Concourse builds
* Capitalize non-english descriptions of commands
* Remove trailing period from command option descriptions
* Remove trailing period from command descriptions
* Capitalized command option descriptions
* Capitalized command descriptions
* Produce manifest with no-hostname attribute
* Support generating manifest with no-hostname
* Replace Ω with Expect, ToNot with NotTo
* Rename Omega to Expect
* Update skip_ssl_validation flag description
* Standardizes on hyphenated locale names
* Make space_quotas test work in Go 1.6
* Make quotas_test test work in Go 1.6
* Make env_var_groups test work in Go 1.6
* Make service_keys_test work in Go 1.6
* Fix bind-route-service min API failure message
* Update create-service-broker creation message
* Remove long name 'force' from delete-route flag
* Update formatting of help text for some commands
* unbind-route-service requires CC >=2.51
* bind-route-service requires CC >=2.51
* update-user-provided-service -r requires CC >=2.51
* Rewrite update_user_provided_service_test
* create-user-provided-service -r requires CC >=2.51
* Remove named return args from cups
* Rewrite create_user_provided_service_test
* Merge pull request #755 from emirozer/master
* Increase StartupTimeout in start test
* RouteServiceBindingRepository includes 'parameters'
* uups does not accept files prefixed with @
* cups does not accept files prefixed with @
* Update help text for curl command
* Remove default method from cf curl -X StringFlag
* fix wrong link in Plugin API doc
* Update curl help text
* curl defaults to POST when body is given
* Update install-plugin help text
* Update remove-plugin-repo help text
* Update untranslated files
* Improve help text of add-plugin-repo
* Remove InputsChan from FakeUI
* Improve help text of repo-plugins
* Improve help text of install-plugin
* Correctly define flags for delete/delete-space
* oauth-token only prints the token
* Remove mapValuesFromPrompt from cups command
* Update tests for update-user-provided-service
* curl -d flag accepts @file
* Add compilers=2 back to .travis.yml
* Extract JSON loader into util package
* bind-route-service takes -c parameters
* Remove unused RouteServiceBinder interface
* Set nodes=4 in travis.yml
* Remove -nodes=4 from bin/test
* Updating a service plan only updates 'public'
* Service auth token commands require API <=2.46
* migrate-service-instances requires api <=2.46
* Add API version req to create-service-broker
* purge-service-offering only takes -p for <=2.46
* Add MaxAPIVersionRequirement
* Remove unused JSON in tests
* update-user-provided-service -p flag accepts @file
* create-user-provided-service -p flag accepts @file
* Remove trailing space from EXAMPLE in usage
* bin/generate-language-resources
* Merge branch 'help2' of https://github.com/fujitsu-cf/cli into fujitsu-cf-help2
* Add space-scoped flag to create-service-broker
* Rewrite create_service_broker_test
* Generate FakeServiceBrokerRepo with counterfeiter
* Make add-plugin-repo test not take 15 seconds
* Fix typo in create-shared-domain usage text again
* replace double-byte colons in main Chinese help text
* Fix typo in create-shared-domain usage text
* bin/generate-language-resources
* Merge branch 'generated-help' of https://github.com/fujitsu-cf/cli into fujitsu-cf-generated-help
* Merge branch 'fix_ja' of https://github.com/tnaoto/cli into tnaoto-fix_ja
* Update README/README-i18n
* Refactor push command a bit
* Add explanatory comment to ProcessPath
* godep save ./... (again)
* godep save ./...
* Bring in missing go-ccapi dependency
* Add V3Apps
* Move passingRequirement definition to suites
* Fix translations to only update the modified messages
* Cleanup test cases for list shared domains
* Add the routing field to 'cf domains' command. Rewrite test cases for domains command
* Enable creating a shared domain with a router group
* fix typo
* Modify Japanese transration
* Push returns an error when given an invalid path
* CheckIfExists ensures path has prefixed slash
* i18n-checkup installs goi18n
* Bind/UnbindRouteService allow --hostname flag
* AuthRepo.Authorize() sets Proxy from environment
* Refactor ssh_code Get() implementation
* Update i18n-checkup
* Regenerate FakeAuthenticationRepository
* Regenerate FakeEndpointRepository
* Revert "WIP merging create route with port"
* Fixing example for unbind-route-services in translation files
* Improving example from usage of the unbind-route-service
* Add -u back to go get of i18n4go
* Login retrieves a max of 50 orgs
* bin/generate-language-resources
* Fix translations
* Add untranslated files
* Print standardized locales when given bad locale
* Fall back to en-us by default
* Rename translation files to what go-i18n expects
* Remove non-translation translations
* Update routes command
* Update unbind_route_service
* Rewrite bind_route_service_test
* Update bind_route_service
* Remove named return args from service_offerings
* Add "repository" to route_service_binding
* Display route services in cf routes command
* Add bind-route-service and unbind-route-service commands
* Add route_service_url flag to update-user-provided-service command.
* Add route_service_url support for create user provided service command.
* HTTP_PROXY -> https_proxy in help template
* cf service includes service tags
* Display X-Cf-Warnings after running commands
* Switch add plugin repo test to use example.com
* On connection refusal include tip to set http proxy
* CopyFiles uses absolute paths
* Update Push actor's handling of paths
* Handle paths that exceed MAX_PATH elsewhere
* WalkAppFiles handles paths that exceed MAX_PATH

## 6.15.0
* Bump version to 6.15.0
* Fix regression in table printer
* Display text where rune len > string len properly
* Remove full stop from files description
* Upload release binaries to v${release_tag}
* Upload release binaries as cf-cli-installer*
* Display non-latin table headers properly
* Remove named return args from CheckIfExists
* CheckIfExists includes path in api call
* files fails for Diego-deployed apps
* Rewrite application/files_test
* Dereference symlinks to app directories
* Keep Alive for non-interactive sessions
* i18n4go -c fixup && bin/generate-language-sources
* Update hostname and path usage
* check-route supports --path option
* Rewrite check_route_test
* push supports --route-path option
* Update push-test to show product of domains/hosts
* Move creation of semver lower in delete-route
* delete-route supports --path option
* Rewrite map_route_test
* Remove unnecessary InputsChan from app_test
* Update de_DE translations
* Display two hyphens for all options
* godep save ./...
* Move simonleung8/flags back into cloudfoundry/cli
* Move prepending of '/' to path into Route Repo
* unmap-route supports --path option
* Update unmap-route
* Rewrite unmap-route test
* updated vcap-dev to cf dev ML
* Re-order create-route requirements
* minRequiredAPIVersion -> requiredVersion
* map-route supports --path option
* Remove map-route Requirements' named return args
* Update MinApiVersionRequirement
* Rewrite map_route_test
* Generate FakeDomainRepository with counterfeiter
* Update create-route hostname help text
* Use ja_JP instead of ja_JA
* Update i18n README to reflect new translations
* Remove extra spaces from some fr_FR entries
* Update create-route examples
* Convert dos format files to unix
* Update translations
* Add --hostname option name for other commands
* Add --hostname option name for create-route
* Set PATH in bin/i18n-checkup
* Fix go vet errors
* Invert conditional in UpdateEndpoint
* Remove named return arguments from UpdateEndpoint
* Don't get i18n4go with -u
* Load core commands later in main
* create-route requires CC 2.36.0+ for --path option
* Remove named return arguments from CreateRoute
* Don't add '/' to path when it's blank
* Update README to reflect move to cloudfoundry/tap
* Update bin/generate-release-notes
* create-route doesn't require prepended / for path
* Update create-route help
* Add context path when listing routes
* Fix help message as per feedback
* Don't use magic empty string in RouteRepository
* Don't construct domain/path with Sprintf
* Move RouteSummary into its own file
* create-route command now takes optional context path
* Update bin/generate-release-notes

## 6.14.1
* Bump version to 6.14.1
* Revert "Merge pull request #718 from cf-routing/list_route_services"
* Revert "Remove extra space in command 'routes' prompt"
* Remove extra space in command 'routes' prompt
* Merge pull request #718 from cf-routing/list_route_services
* Reword help text in `cf push -b` to mention CFF buildpack
* Display route services in cf routes command
* Publish edge installers
* Fix delete-quota force flag text
* Remove developing on windows with powershell note
* Commit i18n_resources.go
* Add missing windows dependencies
* Move GO15VENDOREXPERIMENT lower
* godep save ./...
* Add -p -nodes=4 -randomizeAllSpecs to bin/test
* WalkAppFiles returns error properly
* Remove go vet from bin/get-tools
* Add go fmt changes to fakes
* Update README to reflect switch to Go 1.5
* Add optimization to WalkAppFiles
* pluginAppModel includes more fields
* Export PluginModels
* Remove WalkAppFiles SkipDir optimization
* Skip WalkAppFiles test that expects an error
* Add OS-generic assertions to WalkAppFiles test
* WalkAppFiles skips problematic .cfignored paths
* Don't dot import app_files in test
* Remove named return args from WalkAppFiles
* Update Travis to 1.5.2
* Fix bug in GatherFiles
* Defer close of files in CopyFiles properly
* Add better error when pushing an empty app
* Hoist ProcessPath usage
* Don't use app_files as a local var
* Add error handling to CopyPathToPath use
* Remove copyPathToPath
* Remove presentFiles from GatherFiles
* Remove some private methods from Push actor
* Backfill Push actor tests
* Remove gofileutils from actors/push_test
* Merge pull request #706 from cloudfoundry/shalako-patch-1
* Zipped files on Windows are always 07xx
* Restructure zipper Zip test
* Resource-matched windows files are always 07xx
* GatherFiles does not unzip
* Add ProcessPath to push actor
* Remove empty AfterEach
* Update README.md
* PurgeServiceInstance requires CC API 2.36.0+
* Rewrite purge_service_instance_test
* FakeUI accepts a channel for inputs
* Generate FakeServiceRepository with counterfeiter
* Update MinAPIVersionRequirement
* purge-service-instance: confirmed -> force
* purge-service-instance: Use idiomatic error check
* purge-service-instance: Remove named return args
* fake_registry_command -> fake_command
* Remove diego-ssh/helpers from ssh command
* Remove unused TestHostKeyFingerprint
* Remove SynchronizedBeforeSuite from ssh suite
* bin/test only gets godep if not present
* bin/test doesn't run in a subshell
* Update go vet list in bin/test
* Stop using two copies of fileutils
* godep update gofileutils/fileutils
* Remove usages of CopyReaderToPath
* log_message_queue_old -> loggregator_message_queue
* logs_old_consumer -> log_repository
* Update NewLogMessage helper
* OldLogs -> Logs
* Remove checkForOrgExistence
* Remove unused Noaa code
* Remove unused RunCommandMoreBetter helper
* Move Requirement interface to requirement.go
* Merge pull request #695 from jasonkeene/docs-fix
* Update plugin README
* Move FakeSSHCodeGetter into appropriate location
* Move FakeCommand into appropriate location
* Remove magic from i18n.go
* Sort known locales when printed
* Remove SUPPORTED_LOCALES
* cf/i18n/init.go -> cf/i18n/i18n.go
* Force zh-tw and zh-hk locales to zh-hant
* Improve performance of i18n.Init
* Simplify i18n.Init
* Move `bind-service` fake into appropriate location
* teach command_loader test to ignore fakes
* remove unused fake `FakeApiEndpointSetter`
* Add notes about untested lines in i18n/init
* Fix defers in i18n/init
* Update cf/i18n/init_unix_test
* godep update github.com/nicksnyder/go-i18n/i18n
* Fix build instructions.
* Remove unused logRepo from application/start
* Set BUILD TIME in help as cf.BuiltOnDate always
* Remove Installed-Size from debian install template
* i18n4go -c fixup
* Add proper Installed-Size to debian installers
* Update osx Distribution file
* Show better message for auth ServerErrors
* Remove -p from bin/test ginkgo
* Fix reliance on test pollution in registry_test
* Fix reliance on test pollution in service_keys_test
* godep update github.com/cloudfoundry/jibber_jabber
* Rename SUPPORTED_LOCALES ko_KO -> ko_KR
* Add nodes=4, randomizeAllSpecs to bin/test
* Add -p to bin/build ginkgo
* Rewrite file_download_test
* Add ko_KO i18n support
* Update Contributing section in README
* Remove unused ALL_CAPS const
* findPlan returns error as the last arg
* NoRedirectsErr -> ErrNoRedirects
* Don't assign vars to their defaults
* Don't assign and not use errors
* Run goimports on files that were not goimported
* Don't use underscore as receiver name
* Omit unnecessary 2nd value from range
* Update CHANGELOG.md
* Update README.md
* Update CHANGELOG.md
* v6.14.0 CHANGELOG for plugin API
* Add _osx to osx installer name created by CI
* Add support for prefixed bytes to zipper

## 6.14.0
* Bump version to 6.14.0
* Merge pull request #677 from aminjam/master
* Don't use errors.New(fmt.Sprintf())
* Remove fileutils.Tempdir/file from push command
* Don't repeatedly do the same type assertion
* Output unit 'B' during push when size is in byte size
* Fix test of ApplicationZipper Unzip()
* Add PackageUpdatedAt for cli plugins
* Fix panic printer test
* Update ApplicationZipper Unzip()
* Include empty directories when unzipping archives
* Remove fileutils usage from GatherFiles
* Update crash dialog text and README
* Backfill panic_printer tests
* ignores 10003 error from attempt to add user to org in command 'set-space-role'
* Calculate spacing for both core and plugin commands in 'cf' help
* Do not show plugin command alias in 'cf' help
* use RuneCountInString() instead of len()
* Refactor cli_connection
* Make date reported in --version semver compliant
* Revert "Remove replace-sha"
* Revert "Remove outdated homebrew installer"
* Revert "Fix fallout from 'Remove replace-sha'"
* Revert "Update date format for windows build"
* Update date format for windows build
* Remove outdated homebrew installer
* Remove replace-sha
* Remove strconv.FormatInt(int64(version.Major)
* Replace non-test of config threadsafety with real test
* Use blang/semver
* Update config_data_test
* Update config_repository_test
* assign org role automatically during org creation
* refactor set_org_role command into interface
* Update README.md
* Update README.md
* Revert "Remove unnecessary interface arg"
* Remove terminalUI.Wait()
* Add warning to application start
* Revert 29e491d0f399fe459819d4886ce759def5542963
* Update create-service help text
* Use bytes.NewReader instead of strings.NewReader
* UnsetSpaceRole can unset via username
* Add UnsetSpaceRoleByUsername
* Remove checkSpaceRoleByGuid
* Remove orgRolesPath from users api
* Move api-switching logic into Requirements phase
* Generate fake req. factory with counterfeiter
* Add fallback to user requirement
* Add missing tests for user requirement
* Gateway's newRequest method doesn't return an error
* Remove createUpdateOrDeleteResource
* Remove unnecessary interface arg
* Remove magic number
* UnsetSpaceRole -> UnsetSpaceRoleByGuid
* SetOrgRole -> SetOrgRoleByGuid
* SetOrgRole uses the required user's username
* router-groups command requires RoutingAPIEndpoint
* Generate FakeUserRepository with counterfeiter
* Rewrite api/users_test
* UserRepository UnsetOrgRole -> UnsetOrgRoleByGuid
* UserRepository SetSpaceRole -> SetSpaceRoleByGuid
* Remove errors.NewWithError
* Remove errors.NewWithSlice
* Remove errors.NewWithFmt
* Clean up main.go a bit
* rename var '_' to arg
* provide named argument in test rpc server interface
* use new endpoint when CC support setting space role by name
* Refactor users.go
* Add some error handling to SetOrgRoleByUsername
* Set org role by username
* updated mailing list url from vcap google group to cf-dev
* Document -u flag when go getting go-bindata
* Update upload-binaries-gocd script
* Merge pull request #635 from SrinivasChilveri/Issue_New
* Update upload-binaries-gocd script
* Update upload-binaries-gocd script
* Include cli name and version in binary releases
* Merge pull request #645 from cf-routing/wip-100975070-combined-PR610-PR632
* Merge pull request #611 from utako/purge_service_instance_102318490
* Add purge-service-instance command
* Fix the typo in 'list router-groups' description
* Fix router-group command error handling
* Add cf router-groups command
* Merge pull request #642 from cloudfoundry/multiple-rows-download
* add file extensions, remove 32 bit entries and add reference for 32 bit releases to releases page
* Refresh access token to avoid stale token
* Use -v with i18n4go -c checkup
* Merge pull request #638 from SrinivasChilveri/Isshe_sshcmd
* Fix cf ssh with more than required args
* Fix issues in stacks and stack commands
* Fix the serviceaccess help

## 6.13.0
* Revert "Merge pull request #610 from atulkc/router_group_cli"
* Populate file mode correctly for zip file [#105471590]
* Do not shadow named return value
* Upload file mode under Windows 
* Remove comments related to Noaa [#105524354]
* Update file mode test for multi platform [#105490454]
* Merge pull request #624 from cloudfoundry/integrity-fields-only 
* Merge pull request #622 from cloudfoundry/refactor_user_printing 
* Exclude resource_matches we didn't request [#104364496]
* resource_match requests use only sha1 and size [#104364496]
* Some formatting our build tools wanted
* Tidy and organise userprint package
* Rename package user_printer => userprint 
* not to upload file mode during push under Windows
* skip testing ssh feature in windows
* not to build unix only modules on Windows
* fix filemode test for different platforms
* deps
* Merge pull request #614 from cloudfoundry/zero-users-message-improvement 
* More privatisation. [#63224944]
* Shuffle and correct stuff [#63224944]
* Only pass guid and username to PrintUsers [#63224944]
* Make plugin PrintUsers almost identical [#63224944]
* Strengthen space-users network failure test [#63224944]
* Test error handling for standard space-users call
* Tidy whitespace [#63224944]
* Deterministic ordering from space-users [#63224944]
* Deduplicate call from either side of a branch [#63224944]
* Split space printing into separate types [#63224944]
* Move versioning decision outside of loop [#63224944]
* Iterate over map instead of slice, then map lookup [#63224944]
* Update to use renamed i18n4go binary name [#105200174]
* Merge pull request #610 from atulkc/router_group_cli 
* remove debugging message in create-app-manifest
* better message in 'org-users' when no users found in role [#63224896]
* Rename get-ssh-code to ssh-code [#104476010]
* Merge pull request #609 from cloudfoundry/update-go [#104131294]
* Move to go 1.5.1 [#104131294] 
* Add cf router-groups command [#100975070]
* Remove trailing semicolon in GOPATH [#104131294]
* include file mode during upload file app bits
* command ssh uses one time auth code
* command get-ssh-code
* includes SSHOAuthClient in .cf/config.json
* add command ssh to cf help
* includes file mode during push
* update -t flag usage in ssh command
* SSH command
* add Wildcard in dependency for injecting fakes
* Fix space SSH grouping and capitalisation strings [#102295832]
* Add space-ssh-allowed query [#102295832]
* remove debugging code and unused test 
* remove unused import
* Update dependencies away from code.google.com [#103336616]
* command disallow-space-ssh [#102295832]
* command allow-space-ssh [#102295832]
* allow_ssh field in space model [#102295832]
* ssh-enabled command [#102394414]
* disable-ssh command [#102394414]
* enable-ssh command [#102394414]
* enable_ssh field in models [#102394414]
* push --docker-image help text [#102218860]
* --docker-image for cf push [#102218860]
* adding --health-check-type as full name flag to -u [#101729532]
* godep flags package [#101729532]
* use external flags parsing package with improved features [#101729532]
* add -u to push for health-check-type [#101729532]
* help test for get-health-check and set-health-check [#100320472]
* new get-health-check command [#100320472]
* add HealthCheckType to applicaiton model [#100597578]
* Run i18n4go fixup, which reordered things. [#97265950]
* Merge pull request #607 from cloudfoundry/interactive_plugin_install
* Installer factory constructs the plugin downloader [#97265950]
* Separate files for installers/downloader [#97265950]
* Reshuffles and renames [#97265950]
* Add PluginDownloader abstraction [#97265950]
* Use different var for result of installer.Install [#97265950]
* Go back to context object passing Collapse a dependency into the installer [#97265950]
* More dependency balancing [#97265950]
* Move some deps to specific types [#97265950]
* Pass context bag to plugin installers [#97265950]
* Move payload to structs [#97265950]
* Remove one newline after domain TIPs [#104341944]
* Add more detail to delete-shared-domain error [#104341944]
* Add more detail to delete-domain error [#104341944]
* Add note about generating language resources Fix #529 [#100446378]
* Prepare oauth-token command for plugin execution [#104431292]
* Merge pull request #588 from pishro-cc/master 
* Merge pull request #596 from SrinivasChilveri/Issue_68736518 [#103895494]
* better error reporting with feeded curl data is not enclosed.
* Fix to the delete-domain to fail early if domain is shared delete-domain [#68736518]
* Merge pull request #590 from Zouuup/bug/555 Handles [#103621532] [#101509044]
* Merge pull request #581 from mcowger/master [#103453474] [#103453190]
* Can scale an app to 0 instances [#97749342]
* Update CHANGELOG.md
* typo fixed json test for password containing double quote is now standard
* new test added to test for passwords containing double quotes
* Sanitize now works on passwords containing double quotes
* Change how space tip is presented to user - add quotes. 

##6.12.4
* Merge pull request #589 from cloudfoundry/usage-on-unadorned-push Provide usage on unadorned push
* Provide usage on unadorned push [#103419480]
* Update dependencies away from code.google.com [#103336616]
* diego app not to use noaa for metrics [#103051454]
* Merge pull request #562 from cloudfoundry/BuildpackErrorImprovement Improve error message when app fail to start with "NoAppDetectedError"
* Fix hanging in herd-cats -linux ci script
* Rename the concourse bosh manifest to something revealing aws-vpc.yml has very little meaning or discoverability.
* populate organization name in security-group and security-groups [#102282206]
* create-app-manifest includes command attribute [#102135048]
* Merge pull request #572 from zachgersh/fix-extra-parsing Account for null being passed in extra
* Account for null being passed in extra - fixes #570
* handles cf --version [#102641456]
* Update README.md
* Improve error message when app fail to start with "NoAppDetectedError"
* fix panic in cf marketplace with v1 services
* fix panic in create-app-manifest [#101367528]
* remove inappropriate error message

##6.12.3
* Insert debug messages into CI script
* Merge pull request #544 from cloudfoundry/code-tidy Code tidy
* Merge pull request #523 from zachgersh/master Unmarshal the extra field, get documentation url
* Point to CATS in their new submodule for concourse [#100838442]
* Put job type ahead of architecture in concourse [#100838442]
* Clearer names for cf deployments [#100838442]
* Consistant name for the cli repo Makes it more obvious when you see a path: cli/...[#100838442]
* Unmarshal the extra field, get documentation url 
* add Diego to application model
* Code cleanup: remove unused variables
* Code cleanup: remove orphan functions
* Code cleanup: shadowing reserved word
* update GATS new repo path [#98861144]
* update jibber_jabber repo path [#98861144]
* use go 1.4 to detect symlink file in windows [#75245040]
* do not call GetContainerMetrics() when a diego app is stopped [#98672332]
* Merge pull request #540 from cloudfoundry/use_go_yaml Support yaml '<<' merge type
* remove CodeGangsta dependencies
* complete removal of codegangsta related tests/packages [finishes #97061610 #97061532]
* remove codegangsta from terminal/ui package [#97061610]
* remove orphaned SetApplicationName() in requirements factory
  - func was there to partially support concourrent command calls from plugin. Currently we don't support concurrent calls.
* improve main package's readability [#97061610]
* Clear out codegangsta reference and unused tests in main package [#97061610]
* remove codegangsta from plugin/rpc package - remove deprecated SetTheApp() [#97061610]
* Merge pull request #534 from cloudfoundry/feature/commands-restart-and-create Move commands to new command pattern.
* Move commands to new command pattern. create-user, restart-app-instance [#97061558]
* correct usage text in command space [finishes ##100470938]
* update command_registry test to pass windows
* show executable name instead of CF_NAME in usage help [finishes #100453848]
* handles -v for cf version
* handles usage help without codegangsta
* help command for cmd usage
* handle help menu printing without codegangsta - move cf/app/help into cf/help
  - handles `cf help` / `cf --help` / `cf -h`
* new func Metadatas() in command_registry for returning all metas
* include additional command package in commands_loader
* commands_loader package
* convert commands to non-codegangsta [#97061558]
  - service-key - delete-service-key - update-user-provided-service - unbind-service
* convert commands to non-codegangsta [#97061558] - install-plugin - update-service
* RemoveCommand() in command_registry
* convert commands to non-codegangsta [#97061558]
  - push - disable-service-access - enable-service-access
* convert commands to non-codegangsta [#97061558]
  - copy-source - service-access
* convert commands to non-codegangsta [#97061558]
  - restage - restart - scale - create-space - set-space-role
* convert commands to non-codegangsta [#97061558]
  - bind-service - service-keys - set-space-role
* convert commands to non-codegangsta [#97061558]
  - create-user-provided-service - create-service-key
* BrokerBuilder and PlanBuilder in command_registry.Dependency
* convert commands to non-codegangsta [#97061558]
  - stop - restart
* convert start commands to non-codegangsta [#97061558]
* convert commands to non-codegangsta [#97061558]
  - map-route - create-route
* convert commands to non-codegangsta [#97061558]
  - plugins - uninstall-plugin
* update install-plugin to check conflicts with non-codegangsta commands
* command_registry.CommandExists() returns false for empty string command name
* convert commands to non-codegangsta [#97061558]
  - config - curl - oauth-token - add-plugin-repo - list-plugin-repo - remove-plugin-repo - repo-plugins - update-space-quota
* add plugin_repo.PluginRepo to command_registry dependency
* remove hardcoded version number
* add StringSlice flag feature
* convert commands to non-codegangsta [#97061558]
  - feature-flags - feature-flag - enable-feature-flag - disable-feature-flag
* convert commands to non-codegangsta [#97061558]
  - set-running-environment-variable-group - set-staging-environment-variable-group
* add MaxCommandNameLength() to command_registry
* convert commands to non-codegangsta [#97061558]
  - staging-environment-variable-group
* convert commands to non-codegangsta [#97061558]
  - running-environment-variable-group - bind-running-security=group - running-security-groups - unbind-running-secuirty-group
* convert commands to non-codegangsta [#97061558]
  - bind-security-group - unbind-security-group - staging-seucirty-groups - unbind-staging-security-group - bind-staging-security-group
* update french translation
* convert commands to non-codegangsta [#97061558]
  - create-security-group - delete-security-group - update-security-group
* convert commands to non-codegangsta [#97061558]
  - security-group - security-groups - purge-service-offering
* convert commands to non-codegangsta [#97061558]
  - delete-service-auth-token - update-service-auth-token - create-service-broker - delete-service-broker - rename-service-broker - service-brokers - update-service-broker - migrate-service-instances
* convert commands to non-codegangsta [#97061558]
  - create-service-auth-token - service-auth-tokens - delete-space-quota - set-space-quota - unset-space-quota
* convert space-quotas, space-quota, create-space-quota, update-space-quota to non-codegangsta [#97061558]
* convert rename-service to non-codegangsta [#97061558]
* convert unshare-private-domain to non-codegangsta [#97061558]
* convert share-private-domain to non-codegangsta [#97061558]
* Merge pull request #505 from zhang-hua/bug-93578300 Reduce API calls when CRU operations of service keys
* Merge branch 'story-87481016' of https://github.com/zhang-hua/cli into zhang-hua-story-87481016
* convert share-private-domain to non-codegangsta [#97061558]
* handles 'cf help <command>' for non-codegangsta command
* convert delete-user, set-org-role, unset-org-role to non-codegangsta [#97061558]
* convert delete-service to non-codegangsta [#97061558]
* convert delete-space, rename-space to non-codegangsta [#97061558]
* convert create-quota, delete-quota, update-quota to non-codegangsta [#97061558]
* convert set-quota to non-codegangsta [#97061558]
* convert delete-buildpack, rename-buildpack, quota, quotas to non-codegangsta [#97061558] 
* convert buildpacks, create-buildpack, update-buildpack to non-codegangsta
* Merge pull request #514 from HuaweiTech/hwcf-issue-34 Fix create-app-manifest only includes one host [92530254]
* both godep and travis should use golang v.1.4.2
* make reference to domain test suite for commands to self registered 
  - Godep golang 1.4
* use go v1.4.2 in travis
* convert all check-route, delete-route and delete-orphaned-routes to non-codegangsta [#97061558]
* convert map-route and unmap-route to non-codegangsta
* convert all commands in domain/, rename-org to non-codegangsta [#97061558] 
* convert delete-org to non-codegangsta [#97061558]
* convert create-org to non-codegangsta [#97061558]
* Fix create-app-manifest only includes one host [92530254] 
* SpaceManager and SpaceAuditor should receive 403 [#87481016]
* Reduce API calls when CRU operations of service keys [#93578300]

##6.12.2
* convert create-service to non-codegangsta [#97061558]
* remove used constructor in cmd logs 
* convert marketplace to non-codegangsta [#97061558] 
* add ServiceBuilder to dependency object 
* convert create-app-manifest to non-codegangsta [#97061558] 
* add AppManifest to dependency object 
* convert stack to non-codegangsta [#97061558] 
* convert stacks to non-codegangsta [#97061558] 
* convert unset-env to non-codegangsta [#97061558] 
* convert set-env to non-codegangsta [#97061558] 
* implement skipFlagParsing in flags package [#97061558]
* convert env to non-codegangsta [#97061558] 
* add tip to curl command for api doc url [#98862944]
* convert logs to non-codegangsta [#97061558] 
* convert files to non-codegangsta [#97061558] 
* convert events to non-codegangsta [#97061558] 
* convert rename to non-codegangsta [#97061558] 
* convert delete to non-codegangsta [#97061558]
* cmd passwd converted to non-codegangsta structure [#97061558]
* convert login,logout to non-codegangsta structure [#97061558]
* convert target into non-codegangsta structure [#97061558]
* improve RunCliCommand in testhelper for non-CG command [#97061558]
* change command auth to non codegangsta structure 
* rpc server version check uses new version package [#98664206]
* move version checking methods into utils package [#98664206]
* move NotifyUpdateIfNeeded() into UI package [#98664206]
* Fixed GetMinCliVersion and GetMinApiVersion to work with arbitrary version numbers. [#98664206]
* Populate rpc test server with all plugin API interface 
* Update README.md
* fix bug in plugin API HasAPIEndpoint() 

##6.12.1
* improve method to compare domains of local and redirecting target [98132086]
* Updated config repo fake
* only copy Authorization header when redirecting to same base domain [98132086]
* Revert "Merge pull request #490 from zhang-hua/story-93578300" 
  This reverts commit f449846870ab5fdb360a7345ff83ed73eedfbbfe,
  reversing changes made to 81bf4c37fd40171dd64b48ac57287eb619038fdf.
* security-groups to not use inline-relation-depth to populate spaces model [96033766]
* add spaces_url field to SecurityGroup model [96033766]

##6.12.0 
* Merge pull request #487 from cloudfoundry/96912324-disable-service-access-performance
  - Improve performance of disable-service-access
* Update plugin_examples/README.md 
* Create plugin_examples/DOC.md 
* Merge pull request #490 from zhang-hua/story-93578300
  - Reduce API calls when creating,listing and getting details of service…
* Merge pull request #478 from cloudfoundry/update-empty-tags
  - Allow update service instances with empty tags
* Use expect in test instead of eventually 
* fix race condition in start_test.go 
* fix bug in uninstall-plugin
* add .exe to ignore list in command_factory test
* add needed files for concourse to run
* trigger concourse with cli changes
* enable concourse ci on master branch
* plugin API GetService() [#90442132]
* restructure plugin models file names
* Create unique plugin model for GetServices 
* Create unique plugin model for GetOrgUsers, GetSpaceUsers 
* expand model properties for GetSpace, GetOrg 
* Merge pull request #484 from zhang-hua/list_key_endpoint
  - Change api endpoint for listing service keys
* Create unique plugin model for GetSpace, GetOrg, GetCurrentSpace, GetCurrentOrg
* Create unique plugin model for GetSpaces
* Create unique plugin model for GetOrgs 
* Create unique plugin model for GetApp 
* Create unique plugin model for GetApps
* move command service to non-codegangsta structure [#90442132]
* Reduce API calls when creating,listing and getting details of service keys [#93578300]
  - Leveraging existing API calls in ServiceInstanceRequirement to find service
    instance info by name so that no need to send the same request twice.
* added GetSpaces to api test plugin 
* Merge branch 'improved-service-broker-no-permissions-message' 
* Merge branch 'master' into improved-service-broker-no-permissions-message 
* no translation needed for error text [#95180230]
* Merge pull request #483 from cloudfoundry/service_access_performance
  - improve cf service-access performance
* Merge pull request #470 from cloudfoundry/go14_flake
  - Fix flaky test for go 1.4 where map iteration order is randomized.
* Declare return vars explicitly in func - And return them by name
* Improve performance of disable-service-access - It was making an `async=true` delete request for each
  service_plan_visibility. This meant each delete would take at least 5 seconds due to polling.
- Deleting service plan visibilities does not interact with the broker and can be completed synchronously in ~.5s
- Add new http test matcher for testing empty query strings. [#96912324]
* Refactor to rename SpaceDetails to Space for Plugin API [#97159474]
* Change GetCurrentSpace to use SpaceSummary (vs Space) model [#97159474] 
* Rename OrganizationDetails to Organization in the API Plugin Model [#97159476]
* Change GetCurrentOrg to use OrganizationSummary vs. Organization plugin model structure [#97159476] 
* Add test for GetSpace Plugin API [#97159474] 
* Add getSpace API [#97159474]
* Add plugin API getSpace. [#97159474]
* Change api endpoint for listing service keys [#87481016]
  - CLI should use the endpoint `/v2/service_instances/:fake-guid/service_keys`
    to list service keys instead of using `/v2/service_keys?q=service_instance_guid:fake-guid`
* Backwards compatibility for getCurrentOrg and getCurrentSpace getCurrentOrg returns Organization
getOrg returns OrganizationDetails [#97159474]
getOrgs returns OrganizationSummary
getCurrentSpace returns Space
getSpace returns SpaceDetails
getSpaces returns SpaceSummary
* Change getSpace to be non-CG. Updated some getCurrentSpace which will be reverted [#97159474]
* Make delete service instance as Warn vs. regular Say. make consistent with delete service key 
* Merge pull request #480 from cloudfoundry/missing_service_key_delete
  - Missing service key coloring message from dsk  now matches the coloring from ds
* Reduce service_access API requests: orgs - To map org guids to org names, we make individual requests for each
  org instead of requesting all orgs. [#96912380]
  - This is optimized for the case where there are fewer orgs associated
    with service_plan_visibilities than the total number of org pages.
    This seemed to be the case on all environments we checked.
  - /v2/organizations does not support filtering on a list of org or
    service_plan_visiblility guids, so we have to make separate GETs
- In plan_builder, there are package variables that are used to memoize
  maps. This causes pollution plan_builder tests, so we nil them in test
  setup
* Reduce service_access API requests: service plans [#96912380]
  - Get all service plans in one request instead of a request per service offering

* Reduce service_access API requests: service offerings - Get all service offerings in one request instead of a request per
  broker [#96912380]
* godeps newest noaa package - implement new noaa.Close() method
* Changed the getSpaces API to use SpaceSummary model [#97159474]
* Added space quotas to plugin_model.Organization, fixed plugin API GetCurrentOrg() to work with new org model [ #97159476]
* Add Spaces in plugin API GetOrg() [#97159476]
* Added domains to plugin API GetOrg() [#97159476] 
* Refactor to change Organization to OrganizationSummary for Get Orgs plugin API [#97159476] 
* Add 'org' Plugin API, still needs spaces and domains.. prerefactor for get current org and orgs usage [#97159476]
* Convert 'org' command to non-CG [#97159476] 
* remove windows incompatible language test 
* enable yes for confirmation when lang is not en_US 
* :snowflake: Deflakey-ify the org and space user tests. 
  - Tests were failing in go1.4 due to random org in map.. fixed test to be less brittle
* update vet tool url for travis build 
* Added Services Plugin API [#90441956] 
* Convert services command to non-CG [#90441956]
* Fix up Incorrect Usage i18n in new Plugin APIs [#90440496, #90062486]
* Updated to add the translated string for the usage [#97030456] 
* Implemented the getSpaceUsers plugin API [#90441958]
* Convert spaces-users to non CG [#90441958] 
* Add OrgUsers plugin API [#97030456] 
* Add GetOrgUsers Plugin API [#97030456]
* Add plugin API for Get Org Users [#97030456] 
* Finish convert Org users to non-codegansta cli framework [#97030456]
* Add new plugin test 
* Remove codegansta from Get Org Users [#97030456]
* Missing service key coloring message from dsk  now matches the coloring from ds. - ui type is now `Warn` instead of `Say`
  - Keyword highlight is now switched off [#94220156]
* New plugin api GetSpaces() [#90442002]
* allow command spaces to populate plugin model [#90442002]
* Allow update service instances with empty tags [#96329216]
* convert command `spaces` to non-codegangsta structure [#90442002]
* Fix logic to handle graceful timeout if we cannot talk to log server. 
  - Also make log server connection timeout internally configurable. [#96626036]
* Merge pull request #453 from cloudfoundry/last-operation-timestamps
  - Last operation timestamps
* Updated cf service-access and cf service-brokers so that they only pass through the 403 error, 
  rather than giving specific lookup information. [#91452714]
* Refactor created_at test fixtures [#91240396]
* Updated the CLI to not return a Started date if the service/operation does not have a CreatedAt in it's JSON. [#91240396]
* Add started and updated timestamps to service instance operations [#91240396]
* Merge pull request #465 from cloudfoundry/94892746-service-brokers-403 
  - Expose api errors for service broker commands
* Merge pull request #469 from cloudfoundry/missing_service_key_delete 
  - Display correct error when deleting nonexistent service key
* Merge pull request #472 from cloudfoundry/service_access_performance 
  - Improve performance of enable/disable service access
* Made command_factory_test.go ignore .coverprofile files from running ginkgo in code-coverage mode. [#89585004]
* Update help text for update-service [#72117050]
* Allow `cf app` to display buildpack [#96147958]
* Fixed passing in nil error handler to command_registry [#90652456]
* Merge pull request #463 from cloudfoundry/cli_user_can_provide_tags 
  - Add optional tags to create-service command
* Fix indentation in create-service help text - And rearrange translation files to appease i18n4go
* Add fields to cli msi to show app/publisher name in windows. [#93634720]
* Merge pull request #366 from HuaweiTech/hwcf-issue-15 
  - Fixed error message when there is a mismatch in the order of arguments for create-buildpack
* plugin Api `GetOrgs()` [#90442006]
* enable `orgs` to populate plugin model [#90442006]
* Highlight restage command in uups tip [#96470272]
* convert command `orgs` to non-codegangsta structure [#90442006]
* plugin api GetApps() [#90062486]
* Add Buildpack to cf create-app-manifest [#96041780,91458856]
* Update README.md 
* Update CHANGELOG.md 
* Merge pull request #474 from cloudfoundry/cli_update_service_tags Update user-provided service tags
* Merge pull request #473 from cloudfoundry/i18n-readme-update Update readme with i18n info
* Update error message when plugin file does not exist. [#96267092]
* convert command `apps` to non-codegangsta structure [#90062486]
* add alias support to command_registry [#90062486]
* Update arbitrary params error message [#96313592]
* Merge branch 'master' into cli_update_service_tags Conflicts:
	cf/commands/service/update_service.go
	cf/i18n/resources/de_DE.all.json
	cf/i18n/resources/en_US.all.json
	cf/i18n/resources/es_ES.all.json
	cf/i18n/resources/fr_FR.all.json
	cf/i18n/resources/it_IT.all.json
	cf/i18n/resources/ja_JA.all.json
	cf/i18n/resources/pt_BR.all.json
	cf/i18n/resources/zh_Hans.all.json
	cf/i18n/resources/zh_Hant.all.json
* Update tip for updating UPSIs - UPSIs now propogate their credentials on update, so it is no longer
  necessary to unbind and rebind them. [#96470272]
* Update readme with i18n info 
* Split bind-service usage for easier translation - Improve params example to resemble a bind [#96320118, #72117050]
* Split long usage for update-service [#72117050]
* Update service can pass instance tags - Add ui_helpers/tags_parser.go [#72117050]
* Update service without changing plan works - Fixing a bug where passing arbitrary params without a plan change
  would result in making no changess [#96250704]
* Refactor update service - Plan validation in separate function [#72117050]
* Add optional tags to create-service [#61861194]
* Improve performance of enable/disable service access - Service access commands were embedding org names in service plans, but
  not using them. This resulted in calls to /v2/organizations, which
  would take a long time on environments with many orgs. [#95214984]
* Update help text for update-service [#96313962]
* Merge pull request #440 from xingzhou/service_key_cascade implement the story of delete service instance that has keys
* implement the story of delete service instance that has keys [#92185380]
  https://www.pivotaltracker.com/story/show/92185380
* Fix flaky test for go 1.4 where map iteration order is randomized. [#96235836]
* Display correct error when deleting nonexistent service key [#94220156]
* Merge pull request #452 from cloudfoundry/arbitrary-params-final
  - Arbitrary params for create-service, update-service, bind-service, create-service-key
* Expose api errors for service broker commands - Unless it is a specific case where there was no error but there were
  also no existing service brokers [#94892746]
* Update arbitrary parameter error message - Sometimes it is unclear if the user is intending to provide a file
  path or JSON. Showing the underlying error in these cases can be
  confusing. [#89843658]
* Merge branch 'cmdOutputCapture' 
* update test for non-codegangsta command requirement execution 
* take out unused output capturing method 
* Toggle output to terminal from plugin calls without adding new interface 
* not all calls to non-codegangsta command are from plugin APIs 
* Alternative output capture method - exposes SetOutputBucket() for passing in *[]string as capture bucket
  - passes in nil to disable output capturing.
* Added the changes suggested in the pull request. - Errors no longer overwrite, they bubble up
  - Files are now checked for existance before reading [#89843658]
* Surface error when json from file is invalid - When parsing arbitrary parameters from a file path
  - Only read file contents if we know it's a file [#88670540]
* Merge pull request #365 from HuaweiTech/hwcf-issue-14 Removed as admin.. clause from create-user since it is confusing.
* Added error handling for when diego /instances is up but /noaa is down. [#95483596]
* test should be agnostic to location timezone 
* `GetApp()` plugin api [#90440496]
* plugin model for Application [#90440496]
* new pluginCall field in Command SetDependency() [#90440496]
* convert `app` to non-codegangsta structure [#90440496]
* ShowUsage() to construct cmd usage template [#90440496]
* Merge pull request #443 from xingzhou/service_key_list_newline
  - add a new line before the table of listing keys
* Merge pull request #442 from xingzhou/service_key_detail_newline 
  - add new line before detail output of service key
* move `api` command to new architecture (non-codegangsta) [#90562248]
* flags.String() returns Usage [#90562248]
* command_registry for non-codegangsta command [#90562248]
* Add usage for service key arbitrary params. [#90163332]
* Add more description to bind-service usage - To reflect arbitrary params [#89843654]
* Add detailed usage for update-service - In light of arbitrary params feature [#89843656]
* Remove repeated OPTIONS from create-service [#89843658]
* Add more examples to create-service help file - Arbitrary params examples and description [#89843658]
* User can pass arbitary params during create-service-key Includes code for both json file and raw json [#90163332, #90163330]
* User can pass arbitrary params during  bind-service includes code for both json file and raw json [#89843654, #88670578]
* Do not send async:true in request body for bind-service Two problems: [#92396108]
  1. async flag is a query parameter, not a post body parameter
  2. POST /v2/service_bindings does not respect the async flag anyway
* Add translation for error during update-service with arbitrary params 
* Backfill tests for update-service when sending arbitrary params when they are provided in a file [#88670566]
* user can provide raw JSON when updating a service instance [#89843656]
* add new line before detail output of service key implement story [#94024396]
* add a new line before the table of listing keys implement story [#94026928]
* Fixed error message when there is a mismatch in the order of arguments for create-buildpack. Story in CLI [#82598260].
* Removed as admin.. clause from create-user since it is confusing. Story in CLI [#74893356].

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
* edit cli/plugin/plugin_examples README.md
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
