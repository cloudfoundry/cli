package application_test

import (
	"os"
	"path/filepath"
	"syscall"

	fakeactors "github.com/cloudfoundry/cli/cf/actors/fakes"
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/api/resources"
	testStacks "github.com/cloudfoundry/cli/cf/api/stacks/fakes"
	fakeappfiles "github.com/cloudfoundry/cli/cf/app_files/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/maker"
	testmanifest "github.com/cloudfoundry/cli/testhelpers/manifest"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	testwords "github.com/cloudfoundry/cli/words/generator/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("Push Command", func() {
	var (
		ui                         *testterm.FakeUI
		configRepo                 core_config.Repository
		manifestRepo               *testmanifest.FakeManifestRepository
		starter                    *testcmd.FakeApplicationStarter
		stopper                    *testcmd.FakeApplicationStopper
		serviceBinder              *testcmd.FakeAppBinder
		appRepo                    *testApplication.FakeApplicationRepository
		domainRepo                 *testapi.FakeDomainRepository
		routeRepo                  *testapi.FakeRouteRepository
		stackRepo                  *testStacks.FakeStackRepository
		serviceRepo                *testapi.FakeServiceRepo
		wordGenerator              *testwords.FakeWordGenerator
		requirementsFactory        *testreq.FakeReqFactory
		authRepo                   *testapi.FakeAuthenticationRepository
		actor                      *fakeactors.FakePushActor
		app_files                  *fakeappfiles.FakeAppFiles
		zipper                     *fakeappfiles.FakeZipper
		OriginalCommandStart       command_registry.Command
		OriginalCommandStop        command_registry.Command
		OriginalCommandServiceBind command_registry.Command
		deps                       command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.ManifestRepo = manifestRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo)
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.RepoLocator = deps.RepoLocator.SetStackRepository(stackRepo)
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)
		deps.WordGenerator = wordGenerator
		deps.PushActor = actor
		deps.AppZipper = zipper
		deps.AppFiles = app_files

		//inject fake commands dependencies into registry
		command_registry.Register(starter)
		command_registry.Register(stopper)
		command_registry.Register(serviceBinder)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("push").SetDependency(deps, false))
	}

	BeforeEach(func() {
		manifestRepo = &testmanifest.FakeManifestRepository{}

		starter = &testcmd.FakeApplicationStarter{}
		stopper = &testcmd.FakeApplicationStopper{}
		serviceBinder = &testcmd.FakeAppBinder{}

		//setup fake commands (counterfeiter) to correctly interact with command_registry
		starter.SetDependencyStub = func(_ command_registry.Dependency, _ bool) command_registry.Command {
			return starter
		}
		starter.MetaDataReturns(command_registry.CommandMetadata{Name: "start"})

		stopper.SetDependencyStub = func(_ command_registry.Dependency, _ bool) command_registry.Command {
			return stopper
		}
		stopper.MetaDataReturns(command_registry.CommandMetadata{Name: "stop"})

		appRepo = &testApplication.FakeApplicationRepository{}

		domainRepo = &testapi.FakeDomainRepository{}
		sharedDomain := maker.NewSharedDomainFields(maker.Overrides{"name": "foo.cf-app.com", "guid": "foo-domain-guid"})
		domainRepo.ListDomainsForOrgDomains = []models.DomainFields{sharedDomain}

		//save original command dependences and restore later
		OriginalCommandStart = command_registry.Commands.FindCommand("start")
		OriginalCommandStop = command_registry.Commands.FindCommand("stop")
		OriginalCommandServiceBind = command_registry.Commands.FindCommand("bind-service")

		routeRepo = &testapi.FakeRouteRepository{}
		stackRepo = &testStacks.FakeStackRepository{}
		serviceRepo = &testapi.FakeServiceRepo{}
		authRepo = &testapi.FakeAuthenticationRepository{}
		wordGenerator = new(testwords.FakeWordGenerator)
		wordGenerator.BabbleReturns("laughing-cow")

		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

		zipper = &fakeappfiles.FakeZipper{}
		app_files = &fakeappfiles.FakeAppFiles{}
		actor = &fakeactors.FakePushActor{}

	})

	AfterEach(func() {
		command_registry.Register(OriginalCommandStart)
		command_registry.Register(OriginalCommandStop)
		command_registry.Register(OriginalCommandServiceBind)
	})

	callPush := func(args ...string) bool {
		return testcmd.RunCliCommand("push", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			Expect(callPush()).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(callPush()).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(callPush()).To(BeFalse())
		})

		// yes, we're aware that the args here should probably be provided in a different order
		// erg: app-name -p some/path some-extra-arg
		// but the test infrastructure for parsing args and flags is sorely lacking
		It("fails when provided too many args", func() {
			Expect(callPush("-p", "path", "too-much", "app-name")).To(BeFalse())
		})
	})

	Describe("when pushing a new app", func() {
		BeforeEach(func() {
			appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")

			zipper.ZipReturns(nil)
			zipper.GetZipSizeReturns(9001, nil)
			actor.GatherFilesReturns(nil, true, nil)
			actor.UploadAppReturns(nil)
		})

		It("does not call Zip() when there is no file to be uploaded", func() {
			actor.GatherFilesReturns(nil, false, nil)
			callPush("my-new-app")

			Expect(zipper.ZipCallCount()).To(Equal(0))
		})

		Context("when the default route for the app already exists", func() {
			BeforeEach(func() {
				route := models.Route{}
				route.Guid = "my-route-guid"
				route.Host = "my-new-app"
				route.Domain = domainRepo.ListDomainsForOrgDomains[0]

				routeRepo.FindByHostAndDomainReturns.Route = route
			})

			It("binds to existing routes", func() {
				callPush("my-new-app")

				Expect(routeRepo.CreatedHost).To(BeEmpty())
				Expect(routeRepo.CreatedDomainGuid).To(BeEmpty())
				Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("my-new-app"))
				Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
				Expect(routeRepo.BoundRouteGuid).To(Equal("my-route-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Using", "my-new-app.foo.cf-app.com"},
					[]string{"Binding", "my-new-app.foo.cf-app.com"},
					[]string{"OK"},
				))
			})
		})

		Context("when the default route for the app does not exist", func() {
			BeforeEach(func() {
				routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Org", "couldn't find it")
			})

			It("refreshes the auth token (so fresh)", func() { // so clean
				callPush("fresh-prince")

				Expect(authRepo.RefreshTokenCalled).To(BeTrue())
			})

			Context("when refreshing the auth token fails", func() {
				BeforeEach(func() {
					authRepo.RefreshTokenError = errors.New("I accidentally the UAA")
				})

				It("it displays an error", func() {
					callPush("of-bel-air")

					Expect(ui.Outputs).ToNot(ContainSubstrings(
						[]string{"Error refreshing auth token"},
					))
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"accidentally the UAA"},
					))
				})
			})

			Context("when multiple domains are specified in manifest", func() {

				BeforeEach(func() {
					domainRepo.FindByNameInOrgDomain = []models.DomainFields{
						models.DomainFields{Name: "example1.com", Guid: "example-domain-guid"},
						models.DomainFields{Name: "example2.com", Guid: "example-domain-guid"},
					}

					manifestRepo.ReadManifestReturns.Manifest = multipleDomainsManifest()
				})

				It("creates a route for each domain", func() {
					callPush()

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating", "manifest-host.example1.com"},
						[]string{"OK"},
						[]string{"Binding", "manifest-host.example1.com"},
						[]string{"OK"},
						[]string{"Creating", "manifest-host.example2.com"},
						[]string{"OK"},
						[]string{"Binding", "manifest-host.example2.com"},
						[]string{"OK"},
					))
				})

				It("creates a route for each host on every domains", func() {
					callPush()

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating", "manifest-host.example1.com"},
						[]string{"Binding", "manifest-host.example1.com"},
						[]string{"Creating", "host2.example1.com"},
						[]string{"Binding", "host2.example1.com"},
						[]string{"Creating", "manifest-host.example2.com"},
						[]string{"Binding", "manifest-host.example2.com"},
						[]string{"Creating", "host2.example2.com"},
						[]string{"Binding", "host2.example2.com"},
					))
				})

				It("`-d` from argument will override the domains in manifest", func() {
					callPush("-d", "example1.com")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating", "manifest-host.example1.com"},
						[]string{"OK"},
						[]string{"Binding", "manifest-host.example1.com"},
					))

					Expect(ui.Outputs).ToNot(ContainSubstrings(
						[]string{"Creating", "manifest-host.example2.com"},
					))
				})

			})

			It("creates an app", func() {
				callPush("-t", "111", "my-new-app")
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))

				Expect(*appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
				Expect(*appRepo.CreatedAppParams().SpaceGuid).To(Equal("my-space-guid"))

				Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("my-new-app"))
				Expect(routeRepo.CreatedHost).To(Equal("my-new-app"))
				Expect(routeRepo.CreatedDomainGuid).To(Equal("foo-domain-guid"))
				Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
				Expect(routeRepo.BoundRouteGuid).To(Equal("my-new-app-route-guid"))

				appGuid, _, _ := actor.UploadAppArgsForCall(0)
				Expect(appGuid).To(Equal("my-new-app-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating app", "my-new-app", "my-org", "my-space"},
					[]string{"OK"},
					[]string{"Creating", "my-new-app.foo.cf-app.com"},
					[]string{"OK"},
					[]string{"Binding", "my-new-app.foo.cf-app.com"},
					[]string{"OK"},
					[]string{"Uploading my-new-app"},
					[]string{"OK"},
				))

				Expect(stopper.ApplicationStopCallCount()).To(Equal(0))

				app, orgName, spaceName := starter.ApplicationStartArgsForCall(0)
				Expect(app.Guid).To(Equal(appGuid))
				Expect(app.Name).To(Equal("my-new-app"))
				Expect(orgName).To(Equal(configRepo.OrganizationFields().Name))
				Expect(spaceName).To(Equal(configRepo.SpaceFields().Name))
				Expect(starter.SetStartTimeoutInSecondsArgsForCall(0)).To(Equal(111))
			})

			It("strips special characters when creating a default route", func() {
				callPush("-t", "111", "Tim's 1st-Crazy__app!")
				Expect(*appRepo.CreatedAppParams().Name).To(Equal("Tim's 1st-Crazy__app!"))

				Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("tims-1st-crazy-app"))
				Expect(routeRepo.CreatedHost).To(Equal("tims-1st-crazy-app"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating", "tims-1st-crazy-app.foo.cf-app.com"},
					[]string{"Binding", "tims-1st-crazy-app.foo.cf-app.com"},
				))
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
			})

			It("sets the app params from the flags", func() {
				domainRepo.FindByNameInOrgDomain = []models.DomainFields{
					models.DomainFields{
						Name: "bar.cf-app.com",
						Guid: "bar-domain-guid",
					},
				}
				stackRepo.FindByNameReturns(models.Stack{
					Name: "customLinux",
					Guid: "custom-linux-guid",
				}, nil)

				callPush(
					"-c", "unicorn -c config/unicorn.rb -D",
					"-d", "bar.cf-app.com",
					"-n", "my-hostname",
					"-k", "4G",
					"-i", "3",
					"-m", "2G",
					"-b", "https://github.com/heroku/heroku-buildpack-play.git",
					"-s", "customLinux",
					"-t", "1",
					"--no-start",
					"my-new-app",
				)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Using", "customLinux"},
					[]string{"OK"},
					[]string{"Creating app", "my-new-app"},
					[]string{"OK"},
					[]string{"Creating route", "my-hostname.bar.cf-app.com"},
					[]string{"OK"},
					[]string{"Binding", "my-hostname.bar.cf-app.com", "my-new-app"},
					[]string{"Uploading", "my-new-app"},
					[]string{"OK"},
				))

				Expect(stackRepo.FindByNameArgsForCall(0)).To(Equal("customLinux"))

				Expect(*appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
				Expect(*appRepo.CreatedAppParams().Command).To(Equal("unicorn -c config/unicorn.rb -D"))
				Expect(*appRepo.CreatedAppParams().InstanceCount).To(Equal(3))
				Expect(*appRepo.CreatedAppParams().DiskQuota).To(Equal(int64(4096)))
				Expect(*appRepo.CreatedAppParams().Memory).To(Equal(int64(2048)))
				Expect(*appRepo.CreatedAppParams().StackGuid).To(Equal("custom-linux-guid"))
				Expect(*appRepo.CreatedAppParams().HealthCheckTimeout).To(Equal(1))
				Expect(*appRepo.CreatedAppParams().BuildpackUrl).To(Equal("https://github.com/heroku/heroku-buildpack-play.git"))

				Expect(domainRepo.FindByNameInOrgName).To(Equal("bar.cf-app.com"))
				Expect(domainRepo.FindByNameInOrgGuid).To(Equal("my-org-guid"))

				Expect(routeRepo.CreatedHost).To(Equal("my-hostname"))
				Expect(routeRepo.CreatedDomainGuid).To(Equal("bar-domain-guid"))
				Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
				Expect(routeRepo.BoundRouteGuid).To(Equal("my-hostname-route-guid"))

				appGuid, _, _ := actor.UploadAppArgsForCall(0)
				Expect(appGuid).To(Equal("my-new-app-guid"))

				Expect(starter.ApplicationStartCallCount()).To(Equal(0))
			})

			Context("when there is a shared domain", func() {
				It("creates a route with the shared domain and maps it to the app", func() {
					privateDomain := models.DomainFields{
						Shared: false,
						Name:   "private.cf-app.com",
						Guid:   "private-domain-guid",
					}
					sharedDomain := models.DomainFields{
						Name:   "shared.cf-app.com",
						Shared: true,
						Guid:   "shared-domain-guid",
					}

					domainRepo.ListDomainsForOrgDomains = []models.DomainFields{privateDomain, sharedDomain}

					callPush("-t", "111", "my-new-app")

					Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("my-new-app"))
					Expect(routeRepo.CreatedHost).To(Equal("my-new-app"))
					Expect(routeRepo.CreatedDomainGuid).To(Equal("shared-domain-guid"))
					Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
					Expect(routeRepo.BoundRouteGuid).To(Equal("my-new-app-route-guid"))

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating app", "my-new-app", "my-org", "my-space"},
						[]string{"OK"},
						[]string{"Creating", "my-new-app.shared.cf-app.com"},
						[]string{"OK"},
						[]string{"Binding", "my-new-app.shared.cf-app.com"},
						[]string{"OK"},
						[]string{"Uploading my-new-app"},
						[]string{"OK"},
					))
				})
			})

			Context("when there is no shared domain but there is a private domain in the targeted org", func() {
				It("creates a route with the private domain and maps it to the app", func() {
					privateDomain := models.DomainFields{
						Shared: false,
						Name:   "private.cf-app.com",
						Guid:   "private-domain-guid",
					}

					domainRepo.ListDomainsForOrgDomains = []models.DomainFields{privateDomain}
					appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")

					callPush("-t", "111", "my-new-app")

					Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("my-new-app"))
					Expect(routeRepo.CreatedHost).To(Equal("my-new-app"))
					Expect(routeRepo.CreatedDomainGuid).To(Equal("private-domain-guid"))
					Expect(routeRepo.BoundAppGuid).To(Equal("my-new-app-guid"))
					Expect(routeRepo.BoundRouteGuid).To(Equal("my-new-app-route-guid"))

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating app", "my-new-app", "my-org", "my-space"},
						[]string{"OK"},
						[]string{"Creating", "my-new-app.private.cf-app.com"},
						[]string{"OK"},
						[]string{"Binding", "my-new-app.private.cf-app.com"},
						[]string{"OK"},
						[]string{"Uploading my-new-app"},
						[]string{"OK"},
					))
				})
			})

			Describe("randomized hostnames", func() {
				var manifestApp generic.Map

				BeforeEach(func() {
					manifest := singleAppManifest()
					manifestApp = manifest.Data.Get("applications").([]interface{})[0].(generic.Map)
					manifestApp.Delete("host")
					manifestRepo.ReadManifestReturns.Manifest = manifest
				})

				It("provides a random hostname when the --random-route flag is passed", func() {
					callPush("--random-route", "my-new-app")
					Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("my-new-app-laughing-cow"))
				})

				It("provides a random hostname when the random-route option is set in the manifest", func() {
					manifestApp.Set("random-route", true)

					callPush("my-new-app")

					Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("my-new-app-laughing-cow"))
				})
			})

			It("pushes the contents of the app directory or zip file specified using the -p flag", func() {
				callPush("-p", "../some/path-to/an-app/zip-file", "app-with-path")

				appDir, _ := actor.GatherFilesArgsForCall(0)
				Expect(appDir).To(Equal("../some/path-to/an-app/zip-file"))
			})

			It("pushes the contents of the current working directory by default", func() {
				callPush("app-with-default-path")
				dir, _ := os.Getwd()

				appDir, _ := actor.GatherFilesArgsForCall(0)
				Expect(appDir).To(Equal(dir))
			})

			It("fails when given a bad manifest path", func() {
				manifestRepo.ReadManifestReturns.Manifest = manifest.NewEmptyManifest()
				manifestRepo.ReadManifestReturns.Error = errors.New("read manifest error")

				callPush("-f", "bad/manifest/path")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"read manifest error"},
				))
			})

			It("does not fail when the current working directory does not contain a manifest", func() {
				manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()
				manifestRepo.ReadManifestReturns.Error = syscall.ENOENT
				manifestRepo.ReadManifestReturns.Manifest.Path = ""

				callPush("--no-route", "app-name")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating app", "app-name"},
					[]string{"OK"},
					[]string{"Uploading", "app-name"},
					[]string{"OK"},
				))
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
			})

			It("uses the manifest in the current directory by default", func() {
				manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()
				manifestRepo.ReadManifestReturns.Manifest.Path = "manifest.yml"

				callPush("-p", "some/relative/path")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"Using manifest file", "manifest.yml"}))

				cwd, _ := os.Getwd()
				Expect(manifestRepo.ReadManifestArgs.Path).To(Equal(cwd))
			})

			It("does not use a manifest if the 'no-manifest' flag is passed", func() {
				callPush("--no-route", "--no-manifest", "app-name")

				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"hacker-manifesto"},
				))

				Expect(manifestRepo.ReadManifestArgs.Path).To(Equal(""))
				Expect(*appRepo.CreatedAppParams().Name).To(Equal("app-name"))
			})

			It("pushes an app when provided a manifest with one app defined", func() {
				domainRepo.FindByNameInOrgDomain = []models.DomainFields{
					models.DomainFields{
						Name: "manifest-example.com",
						Guid: "bar-domain-guid",
					},
				}

				manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()

				callPush()

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "manifest-host.manifest-example.com"},
					[]string{"OK"},
					[]string{"Binding", "manifest-host.manifest-example.com"},
					[]string{"manifest-app-name"},
				))

				Expect(*appRepo.CreatedAppParams().Name).To(Equal("manifest-app-name"))
				Expect(*appRepo.CreatedAppParams().Memory).To(Equal(int64(128)))
				Expect(*appRepo.CreatedAppParams().InstanceCount).To(Equal(1))
				Expect(*appRepo.CreatedAppParams().StackName).To(Equal("custom-stack"))
				Expect(*appRepo.CreatedAppParams().BuildpackUrl).To(Equal("some-buildpack"))
				Expect(*appRepo.CreatedAppParams().Command).To(Equal("JAVA_HOME=$PWD/.openjdk JAVA_OPTS=\"-Xss995K\" ./bin/start.sh run"))
				// Expect(actor.UploadedDir).To(Equal(filepath.Clean("some/path/from/manifest"))) TODO: Re-enable this once we develop a strategy

				Expect(*appRepo.CreatedAppParams().EnvironmentVars).To(Equal(map[string]interface{}{
					"PATH": "/u/apps/my-app/bin",
					"FOO":  "baz",
				}))
			})

			It("pushes an app with multiple routes when multiple hosts are provided", func() {
				domainRepo.FindByNameInOrgDomain = []models.DomainFields{
					models.DomainFields{
						Name: "manifest-example.com",
						Guid: "bar-domain-guid",
					},
				}

				manifestRepo.ReadManifestReturns.Manifest = multipleHostManifest()

				callPush()

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "manifest-host-1.manifest-example.com"},
					[]string{"OK"},
					[]string{"Binding", "manifest-host-1.manifest-example.com"},
					[]string{"Creating route", "manifest-host-2.manifest-example.com"},
					[]string{"OK"},
					[]string{"Binding", "manifest-host-2.manifest-example.com"},
					[]string{"manifest-app-name"},
				))
			})

			It("fails when parsing the manifest has errors", func() {
				manifestRepo.ReadManifestReturns.Manifest = &manifest.Manifest{Path: "/some-path/"}
				manifestRepo.ReadManifestReturns.Error = errors.New("buildpack should not be null")

				callPush()

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Error", "reading", "manifest"},
					[]string{"buildpack should not be null"},
				))
			})

			It("does not create a route when provided the --no-route flag", func() {
				domainRepo.FindByNameInOrgDomain = []models.DomainFields{
					models.DomainFields{
						Name: "bar.cf-app.com",
						Guid: "bar-domain-guid",
					},
				}

				callPush("--no-route", "my-new-app")

				Expect(*appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
				Expect(routeRepo.CreatedHost).To(Equal(""))
				Expect(routeRepo.CreatedDomainGuid).To(Equal(""))
			})

			It("maps the root domain route to the app when given the --no-hostname flag", func() {
				domainRepo.ListDomainsForOrgDomains = []models.DomainFields{{
					Name:   "bar.cf-app.com",
					Guid:   "bar-domain-guid",
					Shared: true,
				}}

				routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Org", "uh oh")

				callPush("--no-hostname", "my-new-app")

				Expect(*appRepo.CreatedAppParams().Name).To(Equal("my-new-app"))
				Expect(routeRepo.CreatedHost).To(Equal(""))
				Expect(routeRepo.CreatedDomainGuid).To(Equal("bar-domain-guid"))
			})

			It("Does not create a route when the no-route property is in the manifest", func() {
				workerManifest := singleAppManifest()
				workerManifest.Data.Get("applications").([]interface{})[0].(generic.Map).Set("no-route", true)
				manifestRepo.ReadManifestReturns.Manifest = workerManifest

				callPush("worker-app")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"worker-app", "is a worker", "skipping route creation"}))
				Expect(routeRepo.BoundAppGuid).To(Equal(""))
				Expect(routeRepo.BoundRouteGuid).To(Equal(""))
			})

			It("fails when given an invalid memory limit", func() {
				callPush("-m", "abcM", "my-new-app")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Invalid", "memory limit", "abcM"},
				))
			})

			Context("when a manifest has many apps", func() {
				BeforeEach(func() {
					manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()
				})

				It("pushes each app", func() {
					callPush()

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating", "app1"},
						[]string{"Creating", "app2"},
					))
					Expect(len(appRepo.CreateAppParams)).To(Equal(2))

					firstApp := appRepo.CreateAppParams[0]
					secondApp := appRepo.CreateAppParams[1]
					Expect(*firstApp.Name).To(Equal("app1"))
					Expect(*secondApp.Name).To(Equal("app2"))

					envVars := *firstApp.EnvironmentVars
					Expect(envVars["SOMETHING"]).To(Equal("definitely-something"))

					envVars = *secondApp.EnvironmentVars
					Expect(envVars["SOMETHING"]).To(Equal("nothing"))
				})

				It("pushes a single app when given the name of a single app in the manifest", func() {
					callPush("app2")

					Expect(ui.Outputs).To(ContainSubstrings([]string{"Creating", "app2"}))
					Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Creating", "app1"}))
					Expect(len(appRepo.CreateAppParams)).To(Equal(1))
					Expect(*appRepo.CreateAppParams[0].Name).To(Equal("app2"))
				})

				It("fails when given the name of an app that is not in the manifest", func() {
					callPush("non-existant-app")

					Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
					Expect(len(appRepo.CreateAppParams)).To(Equal(0))
				})
			})
		})
	})

	Describe("re-pushing an existing app", func() {
		var existingApp models.Application

		BeforeEach(func() {
			existingApp = models.Application{}
			existingApp.Name = "existing-app"
			existingApp.Guid = "existing-app-guid"
			existingApp.Command = "unicorn -c config/unicorn.rb -D"
			existingApp.EnvironmentVars = map[string]interface{}{
				"crazy": "pants",
				"FOO":   "NotYoBaz",
				"foo":   "manchu",
			}

			appRepo.ReadReturns.App = existingApp
			appRepo.UpdateAppResult = existingApp
		})

		It("resets the app's buildpack when the -b flag is provided as 'default'", func() {
			callPush("-b", "default", "existing-app")
			Expect(*appRepo.UpdateParams.BuildpackUrl).To(Equal(""))
		})

		It("resets the app's command when the -c flag is provided as 'default'", func() {
			callPush("-c", "default", "existing-app")
			Expect(*appRepo.UpdateParams.Command).To(Equal(""))
		})

		It("resets the app's buildpack when the -b flag is provided as 'null'", func() {
			callPush("-b", "null", "existing-app")
			Expect(*appRepo.UpdateParams.BuildpackUrl).To(Equal(""))
		})

		It("resets the app's command when the -c flag is provided as 'null'", func() {
			callPush("-c", "null", "existing-app")
			Expect(*appRepo.UpdateParams.Command).To(Equal(""))
		})

		It("merges env vars from the manifest with those from the server", func() {
			manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()

			callPush("existing-app")

			updatedAppEnvVars := *appRepo.UpdateParams.EnvironmentVars
			Expect(updatedAppEnvVars["crazy"]).To(Equal("pants"))
			Expect(updatedAppEnvVars["FOO"]).To(Equal("baz"))
			Expect(updatedAppEnvVars["foo"]).To(Equal("manchu"))
			Expect(updatedAppEnvVars["PATH"]).To(Equal("/u/apps/my-app/bin"))
		})

		It("stops the app, achieving a full-downtime deploy!", func() {
			appRepo.UpdateAppResult = existingApp

			callPush("existing-app")

			app, orgName, spaceName := stopper.ApplicationStopArgsForCall(0)
			Expect(app.Guid).To(Equal(existingApp.Guid))
			Expect(app.Name).To(Equal("existing-app"))
			Expect(orgName).To(Equal(configRepo.OrganizationFields().Name))
			Expect(spaceName).To(Equal(configRepo.SpaceFields().Name))

			appGuid, _, _ := actor.UploadAppArgsForCall(0)
			Expect(appGuid).To(Equal(existingApp.Guid))
		})

		It("does not stop the app when it is already stopped", func() {
			existingApp.State = "stopped"
			appRepo.ReadReturns.App = existingApp
			appRepo.UpdateAppResult = existingApp

			callPush("existing-app")

			Expect(stopper.ApplicationStopCallCount()).To(Equal(0))
		})

		It("updates the app", func() {
			existingRoute := models.RouteSummary{}
			existingRoute.Host = "existing-app"

			existingApp.Routes = []models.RouteSummary{existingRoute}
			appRepo.ReadReturns.App = existingApp
			appRepo.UpdateAppResult = existingApp

			stackRepo.FindByNameReturns(models.Stack{
				Name: "differentStack",
				Guid: "differentStack-guid",
			}, nil)

			callPush(
				"-c", "different start command",
				"-i", "10",
				"-m", "1G",
				"-b", "https://github.com/heroku/heroku-buildpack-different.git",
				"-s", "differentStack",
				"existing-app",
			)

			Expect(appRepo.UpdateAppGuid).To(Equal(existingApp.Guid))
			Expect(*appRepo.UpdateParams.Command).To(Equal("different start command"))
			Expect(*appRepo.UpdateParams.InstanceCount).To(Equal(10))
			Expect(*appRepo.UpdateParams.Memory).To(Equal(int64(1024)))
			Expect(*appRepo.UpdateParams.BuildpackUrl).To(Equal("https://github.com/heroku/heroku-buildpack-different.git"))
			Expect(*appRepo.UpdateParams.StackGuid).To(Equal("differentStack-guid"))
		})

		It("re-uploads the app", func() {
			callPush("existing-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Uploading", "existing-app"},
				[]string{"OK"},
			))
		})

		Describe("when the app has a route bound", func() {
			BeforeEach(func() {
				domain := models.DomainFields{
					Name:   "example.com",
					Guid:   "domain-guid",
					Shared: true,
				}

				domainRepo.ListDomainsForOrgDomains = []models.DomainFields{domain}
				routeRepo.FindByHostAndDomainReturns.Route = models.Route{
					Host:   "existing-app",
					Domain: domain,
				}

				existingApp.Routes = []models.RouteSummary{models.RouteSummary{
					Guid:   "existing-route-guid",
					Host:   "existing-app",
					Domain: domain,
				}}

				appRepo.ReadReturns.App = existingApp
				appRepo.UpdateAppResult = existingApp
			})

			It("uses the existing route when an app already has it bound", func() {
				callPush("-d", "example.com", "existing-app")

				Expect(routeRepo.CreatedHost).To(Equal(""))
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Creating route"}))
				Expect(ui.Outputs).To(ContainSubstrings([]string{"Using route", "existing-app", "example.com"}))
			})

			Context("and no route-related flags are given", func() {
				Context("and there is no route in the manifest", func() {
					It("does not add a route to the app", func() {
						callPush("existing-app")

						appGuid, _, _ := actor.UploadAppArgsForCall(0)
						Expect(appGuid).To(Equal("existing-app-guid"))
						Expect(domainRepo.FindByNameInOrgName).To(Equal(""))
						Expect(routeRepo.FindByHostAndDomainCalledWith.Domain.Name).To(Equal(""))
						Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal(""))
						Expect(routeRepo.CreatedHost).To(Equal(""))
						Expect(routeRepo.CreatedDomainGuid).To(Equal(""))
					})
				})

				Context("and there is a route in the manifest", func() {
					BeforeEach(func() {
						manifestRepo.ReadManifestReturns.Manifest = existingAppManifest()

						routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Org", "uh oh")

						domainRepo.FindByNameInOrgDomain = []models.DomainFields{
							models.DomainFields{Name: "example.com", Guid: "example-domain-guid"},
						}
					})

					It("adds the route", func() {
						callPush("existing-app")
						Expect(routeRepo.CreatedHost).To(Equal("new-manifest-host"))
					})
				})
			})

			It("creates and binds a route when a different domain is specified", func() {
				routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Org", "existing-app.newdomain.com")
				domainRepo.FindByNameInOrgDomain = []models.DomainFields{
					models.DomainFields{Guid: "domain-guid", Name: "newdomain.com"},
				}

				callPush("-d", "newdomain.com", "existing-app")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "existing-app.newdomain.com"},
					[]string{"OK"},
					[]string{"Binding", "existing-app.newdomain.com"},
				))

				Expect(domainRepo.FindByNameInOrgName).To(Equal("newdomain.com"))
				Expect(domainRepo.FindByNameInOrgGuid).To(Equal("my-org-guid"))
				Expect(routeRepo.FindByHostAndDomainCalledWith.Domain.Name).To(Equal("newdomain.com"))
				Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("existing-app"))
				Expect(routeRepo.CreatedHost).To(Equal("existing-app"))
				Expect(routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
			})

			It("creates and binds a route when a different hostname is specified", func() {
				routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Org", "new-host.newdomain.com")

				callPush("-n", "new-host", "existing-app")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "new-host.example.com"},
					[]string{"OK"},
					[]string{"Binding", "new-host.example.com"},
				))

				Expect(routeRepo.FindByHostAndDomainCalledWith.Domain.Name).To(Equal("example.com"))
				Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal("new-host"))
				Expect(routeRepo.CreatedHost).To(Equal("new-host"))
				Expect(routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
			})

			It("removes the route when the --no-route flag is given", func() {
				callPush("--no-route", "existing-app")

				appGuid, _, _ := actor.UploadAppArgsForCall(0)
				Expect(appGuid).To(Equal("existing-app-guid"))
				Expect(domainRepo.FindByNameInOrgName).To(Equal(""))
				Expect(routeRepo.FindByHostAndDomainCalledWith.Domain.Name).To(Equal(""))
				Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal(""))
				Expect(routeRepo.CreatedHost).To(Equal(""))
				Expect(routeRepo.CreatedDomainGuid).To(Equal(""))
				Expect(routeRepo.UnboundRouteGuid).To(Equal("existing-route-guid"))
				Expect(routeRepo.UnboundAppGuid).To(Equal("existing-app-guid"))
			})

			It("binds the root domain route to an app with a pre-existing route when the --no-hostname flag is given", func() {
				routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Org", "existing-app.example.com")

				callPush("--no-hostname", "existing-app")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "example.com"},
					[]string{"OK"},
					[]string{"Binding", "example.com"},
				))
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"existing-app.example.com"}))

				Expect(routeRepo.FindByHostAndDomainCalledWith.Domain.Name).To(Equal("example.com"))
				Expect(routeRepo.FindByHostAndDomainCalledWith.Host).To(Equal(""))
				Expect(routeRepo.CreatedHost).To(Equal(""))
				Expect(routeRepo.CreatedDomainGuid).To(Equal("domain-guid"))
			})
		})
	})

	Describe("service instances", func() {
		BeforeEach(func() {
			serviceRepo.FindInstanceByNameMap = generic.NewMap(map[interface{}]interface{}{
				"global-service": maker.NewServiceInstance("global-service"),
				"app1-service":   maker.NewServiceInstance("app1-service"),
				"app2-service":   maker.NewServiceInstance("app2-service"),
			})

			manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()
		})

		Context("when the service is not bound", func() {
			BeforeEach(func() {
				appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")
			})

			It("binds service instances to the app", func() {
				callPush()
				Expect(len(serviceBinder.AppsToBind)).To(Equal(4))
				Expect(serviceBinder.AppsToBind[0].Name).To(Equal("app1"))
				Expect(serviceBinder.AppsToBind[1].Name).To(Equal("app1"))
				Expect(serviceBinder.InstancesToBindTo[0].Name).To(Equal("app1-service"))
				Expect(serviceBinder.InstancesToBindTo[1].Name).To(Equal("global-service"))

				Expect(serviceBinder.AppsToBind[2].Name).To(Equal("app2"))
				Expect(serviceBinder.AppsToBind[3].Name).To(Equal("app2"))
				Expect(serviceBinder.InstancesToBindTo[2].Name).To(Equal("app2-service"))
				Expect(serviceBinder.InstancesToBindTo[3].Name).To(Equal("global-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating", "app1"},
					[]string{"OK"},
					[]string{"Binding service", "app1-service", "app1", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"Binding service", "global-service", "app1", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"Creating", "app2"},
					[]string{"OK"},
					[]string{"Binding service", "app2-service", "app2", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"Binding service", "global-service", "app2", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
			})
		})

		Context("when the app is already bound to the service", func() {
			BeforeEach(func() {
				appRepo.ReadReturns.App = maker.NewApp(maker.Overrides{})
				serviceBinder.BindApplicationReturns.Error = errors.NewHttpError(500, "90003", "it don't work")
			})

			It("gracefully continues", func() {
				callPush()
				Expect(len(serviceBinder.AppsToBind)).To(Equal(4))
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when the service instance can't be found", func() {
			BeforeEach(func() {
				//				routeRepo.FindByHostAndDomainReturns.Error = errors.new("can't find service instance")
				serviceRepo.FindInstanceByNameErr = true
				manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()
			})

			It("fails with an error", func() {
				callPush()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Could not find service", "app1-service", "app1"},
				))
			})
		})

	})

	Describe("checking for bad flags", func() {
		It("fails when a non-numeric start timeout is given", func() {
			appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")

			callPush("-t", "FooeyTimeout", "my-new-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Invalid", "timeout", "FooeyTimeout"},
			))
		})
	})

	Describe("displaying information about files being uploaded", func() {
		It("displays information about the files being uploaded", func() {
			app_files.CountFilesReturns(11)
			zipper.ZipReturns(nil)
			zipper.GetZipSizeReturns(6100000, nil)
			actor.GatherFilesReturns([]resources.AppFileResource{resources.AppFileResource{Path: "path/to/app"}, resources.AppFileResource{Path: "bar"}}, true, nil)

			curDir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			callPush("appName")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Uploading", curDir},
				[]string{"5.8M", "11 files"},
			))
		})
	})

	It("fails when the app can't be uploaded", func() {
		actor.UploadAppReturns(errors.New("Boom!"))

		callPush("app")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Uploading"},
			[]string{"FAILED"},
		))
	})

	Describe("when binding the route fails", func() {
		BeforeEach(func() {
			routeRepo.FindByHostAndDomainReturns.Route.Host = "existing-app"
			routeRepo.FindByHostAndDomainReturns.Route.Domain = models.DomainFields{Name: "foo.cf-app.com"}
		})

		It("suggests using 'random-route' if the default route is taken", func() {
			routeRepo.BindErr = errors.NewHttpError(400, errors.INVALID_RELATION, "The URL not available")

			callPush("existing-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"existing-app.foo.cf-app.com", "already in use"},
				[]string{"TIP", "random-route"},
			))
		})

		It("does not suggest using 'random-route' for other failures", func() {
			routeRepo.BindErr = errors.NewHttpError(500, "some-code", "exception happened")

			callPush("existing-app")

			Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"TIP", "random-route"}))
		})
	})

	It("fails when neither a manifest nor a name is given", func() {
		manifestRepo.ReadManifestReturns.Error = errors.New("No such manifest")
		callPush()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"Manifest file is not found"},
		))
	})
})

func existingAppManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Path: "manifest.yml",
		Data: generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"name":      "manifest-app-name",
					"memory":    "128MB",
					"instances": 1,
					"host":      "new-manifest-host",
					"domain":    "example.com",
					"stack":     "custom-stack",
					"timeout":   360,
					"buildpack": "some-buildpack",
					"command":   `JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run`,
					"path":      filepath.Clean("some/path/from/manifest"),
					"env": generic.NewMap(map[interface{}]interface{}{
						"FOO":  "baz",
						"PATH": "/u/apps/my-app/bin",
					}),
				}),
			},
		}),
	}
}

func multipleHostManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Path: "manifest.yml",
		Data: generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"name":      "manifest-app-name",
					"memory":    "128MB",
					"instances": 1,
					"hosts":     []interface{}{"manifest-host-1", "manifest-host-2"},
					"domain":    "manifest-example.com",
					"stack":     "custom-stack",
					"timeout":   360,
					"buildpack": "some-buildpack",
					"command":   `JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run`,
					"path":      filepath.Clean("some/path/from/manifest"),
					"env": generic.NewMap(map[interface{}]interface{}{
						"FOO":  "baz",
						"PATH": "/u/apps/my-app/bin",
					}),
				}),
			},
		}),
	}
}

func multipleDomainsManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Path: "manifest.yml",
		Data: generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"name":      "manifest-app-name",
					"memory":    "128MB",
					"instances": 1,
					"host":      "manifest-host",
					"hosts":     []interface{}{"host2"},
					"domains":   []interface{}{"example1.com", "example2.com"},
					"stack":     "custom-stack",
					"timeout":   360,
					"buildpack": "some-buildpack",
					"command":   `JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run`,
					"path":      filepath.Clean("some/path/from/manifest"),
					"env": generic.NewMap(map[interface{}]interface{}{
						"FOO":  "baz",
						"PATH": "/u/apps/my-app/bin",
					}),
				}),
			},
		}),
	}
}

func singleAppManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Path: "manifest.yml",
		Data: generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"name":      "manifest-app-name",
					"memory":    "128MB",
					"instances": 1,
					"host":      "manifest-host",
					"domain":    "manifest-example.com",
					"stack":     "custom-stack",
					"timeout":   360,
					"buildpack": "some-buildpack",
					"command":   `JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run`,
					"path":      filepath.Clean("some/path/from/manifest"),
					"env": generic.NewMap(map[interface{}]interface{}{
						"FOO":  "baz",
						"PATH": "/u/apps/my-app/bin",
					}),
				}),
			},
		}),
	}
}

func manifestWithServicesAndEnv() *manifest.Manifest {
	return &manifest.Manifest{
		Data: generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"name":     "app1",
					"services": []interface{}{"app1-service", "global-service"},
					"env": generic.NewMap(map[interface{}]interface{}{
						"SOMETHING": "definitely-something",
					}),
				}),
				generic.NewMap(map[interface{}]interface{}{
					"name":     "app2",
					"services": []interface{}{"app2-service", "global-service"},
					"env": generic.NewMap(map[interface{}]interface{}{
						"SOMETHING": "nothing",
					}),
				}),
			},
		}),
	}
}
