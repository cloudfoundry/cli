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
	appCmdFakes "github.com/cloudfoundry/cli/cf/commands/application/fakes"
	serviceCmdFakes "github.com/cloudfoundry/cli/cf/commands/service/fakes"
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
		starter                    *appCmdFakes.FakeApplicationStarter
		stopper                    *appCmdFakes.FakeApplicationStopper
		serviceBinder              *serviceCmdFakes.FakeAppBinder
		appRepo                    *testApplication.FakeApplicationRepository
		domainRepo                 *testapi.FakeDomainRepository
		routeRepo                  *testapi.FakeRouteRepository
		stackRepo                  *testStacks.FakeStackRepository
		serviceRepo                *testapi.FakeServiceRepository
		wordGenerator              *testwords.FakeWordGenerator
		requirementsFactory        *testreq.FakeReqFactory
		authRepo                   *testapi.FakeAuthenticationRepository
		actor                      *fakeactors.FakePushActor
		appfiles                   *fakeappfiles.FakeAppFiles
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
		deps.AppFiles = appfiles

		//inject fake commands dependencies into registry
		command_registry.Register(starter)
		command_registry.Register(stopper)
		command_registry.Register(serviceBinder)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("push").SetDependency(deps, false))
	}

	BeforeEach(func() {
		manifestRepo = &testmanifest.FakeManifestRepository{}

		starter = &appCmdFakes.FakeApplicationStarter{}
		stopper = &appCmdFakes.FakeApplicationStopper{}
		serviceBinder = &serviceCmdFakes.FakeAppBinder{}

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
		domainRepo.ListDomainsForOrgStub = func(orgGuid string, cb func(models.DomainFields) bool) error {
			cb(sharedDomain)
			return nil
		}

		domainRepo.FirstOrDefaultStub = func(orgGuid string, name *string) (models.DomainFields, error) {
			if name == nil {
				var foundDomain *models.DomainFields
				domainRepo.ListDomainsForOrg(orgGuid, func(domain models.DomainFields) bool {
					foundDomain = &domain
					return !domain.Shared
				})

				if foundDomain == nil {
					return models.DomainFields{}, errors.New("Could not find a default domain")
				}

				return *foundDomain, nil
			}

			return domainRepo.FindByNameInOrg(*name, orgGuid)
		}

		//save original command dependences and restore later
		OriginalCommandStart = command_registry.Commands.FindCommand("start")
		OriginalCommandStop = command_registry.Commands.FindCommand("stop")
		OriginalCommandServiceBind = command_registry.Commands.FindCommand("bind-service")

		routeRepo = &testapi.FakeRouteRepository{}
		routeRepo.CreateStub = func(host string, domain models.DomainFields, path string) (models.Route, error) {
			// This never returns an error, which means it isn't tested.
			// This is copied from the old route repo fake.
			route := models.Route{}
			route.Guid = host + "-route-guid"
			route.Domain = domain
			route.Host = host
			route.Path = path

			return route, nil
		}

		stackRepo = &testStacks.FakeStackRepository{}
		serviceRepo = &testapi.FakeServiceRepository{}
		authRepo = &testapi.FakeAuthenticationRepository{}
		wordGenerator = new(testwords.FakeWordGenerator)
		wordGenerator.BabbleReturns("random-host")

		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()

		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
			MinAPIVersionSuccess: true,
		}

		zipper = &fakeappfiles.FakeZipper{}
		appfiles = &fakeappfiles.FakeAppFiles{}
		appfiles.AppFilesInDirReturns([]models.AppFileFields{
			{
				Path: "some-path",
			},
		}, nil)
		actor = &fakeactors.FakePushActor{}
		actor.ProcessPathStub = func(dirOrZipFile string, f func(string)) error {
			f(dirOrZipFile)
			return nil
		}
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

		Context("when the CC API version is too low", func() {
			BeforeEach(func() {
				requirementsFactory.MinAPIVersionSuccess = false
			})

			It("fails when provided the --route-path option", func() {
				Expect(callPush("--route-path", "the-path", "app-name")).To(BeFalse())
			})
		})

		Context("when the CC API version is not too low", func() {
			BeforeEach(func() {
				requirementsFactory.MinAPIVersionSuccess = true
			})

			It("does not fail when provided the --route-path option", func() {
				Expect(callPush("--route-path", "the-path", "app-name")).To(BeTrue())
			})
		})
	})

	Describe("when pushing a new app", func() {
		BeforeEach(func() {
			appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))
			appRepo.CreateStub = func(params models.AppParams) (models.Application, error) {
				a := models.Application{}
				a.Guid = *params.Name + "-guid"
				a.Name = *params.Name
				a.State = "stopped"

				return a, nil
			}

			zipper.ZipReturns(nil)
			zipper.GetZipSizeReturns(9001, nil)
			actor.GatherFilesReturns(nil, true, nil)
			actor.UploadAppReturns(nil)
		})

		Context("when the default route for the app already exists", func() {
			BeforeEach(func() {
				route := models.Route{}
				route.Guid = "my-route-guid"
				route.Host = "app-name"
				route.Domain = maker.NewSharedDomainFields(maker.Overrides{"name": "foo.cf-app.com", "guid": "foo-domain-guid"})

				routeRepo.FindReturns(route, nil)
			})

			It("notifies users about the error actor.GatherFiles() returns", func() {
				actor.GatherFilesReturns([]resources.AppFileResource{}, false, errors.New("failed to get file mode"))

				callPush("app-name")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"failed to get file mode"},
				))
			})

			It("binds to existing routes", func() {
				callPush("app-name")

				Expect(routeRepo.CreateCallCount()).To(BeZero())

				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, _, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("app-name"))

				Expect(routeRepo.BindCallCount()).To(Equal(1))
				boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
				Expect(boundAppGUID).To(Equal("app-name-guid"))
				Expect(boundRouteGUID).To(Equal("my-route-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Using", "app-name.foo.cf-app.com"},
					[]string{"Binding", "app-name.foo.cf-app.com"},
					[]string{"OK"},
				))
			})
		})

		Context("when the default route for the app does not exist", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "couldn't find it"))
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
					domainRepo.FindByNameInOrgStub = func(name string, owningOrgGuid string) (models.DomainFields, error) {
						return map[string]models.DomainFields{
							"example1.com": {Name: "example1.com", Guid: "example-domain-guid"},
							"example2.com": {Name: "example2.com", Guid: "example-domain-guid"},
						}[name], nil
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
						[]string{"Creating", "host2.example1.com"},
						[]string{"OK"},
						[]string{"Binding", "host2.example1.com"},
						[]string{"OK"},
						[]string{"Creating", "host2.example2.com"},
						[]string{"OK"},
						[]string{"Binding", "host2.example2.com"},
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
				callPush("-t", "111", "app-name")
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))

				params := appRepo.CreateArgsForCall(0)
				Expect(*params.Name).To(Equal("app-name"))
				Expect(*params.SpaceGuid).To(Equal("my-space-guid"))

				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, _, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("app-name"))

				Expect(routeRepo.CreateCallCount()).To(Equal(1))
				createdHost, createdDomainFields, createdPath := routeRepo.CreateArgsForCall(0)
				Expect(createdHost).To(Equal("app-name"))
				Expect(createdDomainFields.Guid).To(Equal("foo-domain-guid"))
				Expect(createdPath).To(BeEmpty())

				Expect(routeRepo.BindCallCount()).To(Equal(1))
				boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
				Expect(boundAppGUID).To(Equal("app-name-guid"))
				Expect(boundRouteGUID).To(Equal("app-name-route-guid"))

				appGuid, _, _ := actor.UploadAppArgsForCall(0)
				Expect(appGuid).To(Equal("app-name-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating app", "app-name", "my-org", "my-space"},
					[]string{"OK"},
					[]string{"Creating", "app-name.foo.cf-app.com"},
					[]string{"OK"},
					[]string{"Binding", "app-name.foo.cf-app.com"},
					[]string{"OK"},
					[]string{"Uploading app-name"},
					[]string{"OK"},
				))

				Expect(stopper.ApplicationStopCallCount()).To(Equal(0))

				app, orgName, spaceName := starter.ApplicationStartArgsForCall(0)
				Expect(app.Guid).To(Equal(appGuid))
				Expect(app.Name).To(Equal("app-name"))
				Expect(orgName).To(Equal(configRepo.OrganizationFields().Name))
				Expect(spaceName).To(Equal(configRepo.SpaceFields().Name))
				Expect(starter.SetStartTimeoutInSecondsArgsForCall(0)).To(Equal(111))
			})

			It("strips special characters when creating a default route", func() {
				callPush("-t", "111", "app!name")

				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, _, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("appname"))
				Expect(routeRepo.CreateCallCount()).To(Equal(1))
				createdHost, _, _ := routeRepo.CreateArgsForCall(0)
				Expect(createdHost).To(Equal("appname"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating", "appname.foo.cf-app.com"},
					[]string{"Binding", "appname.foo.cf-app.com"},
				))
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
			})

			It("sets the app params from the flags", func() {
				domainRepo.FindByNameInOrgReturns(models.DomainFields{
					Name: "bar.cf-app.com",
					Guid: "bar-domain-guid",
				}, nil)
				stackRepo.FindByNameReturns(models.Stack{
					Name: "customLinux",
					Guid: "custom-linux-guid",
				}, nil)

				callPush(
					"-c", "unicorn -c config/unicorn.rb -D",
					"-d", "bar.cf-app.com",
					"-n", "my-hostname",
					"--route-path", "my-route-path",
					"-k", "4G",
					"-i", "3",
					"-m", "2G",
					"-b", "https://github.com/heroku/heroku-buildpack-play.git",
					"-s", "customLinux",
					"-t", "1",
					"-u", "port",
					"--no-start",
					"app-name",
				)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Using", "customLinux"},
					[]string{"OK"},
					[]string{"Creating app", "app-name"},
					[]string{"OK"},
					[]string{"Creating route", "my-hostname.bar.cf-app.com/my-route-path"},
					[]string{"OK"},
					[]string{"Binding", "my-hostname.bar.cf-app.com/my-route-path", "app-name"},
					[]string{"Uploading", "app-name"},
					[]string{"OK"},
				))

				Expect(stackRepo.FindByNameArgsForCall(0)).To(Equal("customLinux"))

				params := appRepo.CreateArgsForCall(0)
				Expect(*params.Name).To(Equal("app-name"))
				Expect(*params.Command).To(Equal("unicorn -c config/unicorn.rb -D"))
				Expect(*params.InstanceCount).To(Equal(3))
				Expect(*params.DiskQuota).To(Equal(int64(4096)))
				Expect(*params.Memory).To(Equal(int64(2048)))
				Expect(*params.StackGuid).To(Equal("custom-linux-guid"))
				Expect(*params.HealthCheckTimeout).To(Equal(1))
				Expect(*params.HealthCheckType).To(Equal("port"))
				Expect(*params.BuildpackUrl).To(Equal("https://github.com/heroku/heroku-buildpack-play.git"))

				name, owningOrgGuid := domainRepo.FindByNameInOrgArgsForCall(0)
				Expect(name).To(Equal("bar.cf-app.com"))
				Expect(owningOrgGuid).To(Equal("my-org-guid"))

				Expect(routeRepo.CreateCallCount()).To(Equal(1))
				createdHost, createdDomainFields, createdPath := routeRepo.CreateArgsForCall(0)
				Expect(createdHost).To(Equal("my-hostname"))
				Expect(createdDomainFields.Guid).To(Equal("bar-domain-guid"))
				Expect(createdPath).To(Equal("my-route-path"))

				Expect(routeRepo.BindCallCount()).To(Equal(1))
				boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
				Expect(boundAppGUID).To(Equal("app-name-guid"))
				Expect(boundRouteGUID).To(Equal("my-hostname-route-guid"))

				appGuid, _, _ := actor.UploadAppArgsForCall(0)
				Expect(appGuid).To(Equal("app-name-guid"))

				Expect(starter.ApplicationStartCallCount()).To(Equal(0))
			})

			Context("when pushing a docker image with --docker-image or -o", func() {
				It("sets diego to true", func() {
					callPush("testApp", "--docker-image", "sample/dockerImage")

					Expect(appRepo.CreateCallCount()).To(Equal(1))
					params := appRepo.CreateArgsForCall(0)
					Expect(*params.Diego).To(BeTrue())
				})

				It("sets docker_image", func() {
					callPush("testApp", "-o", "sample/dockerImage")

					params := appRepo.CreateArgsForCall(0)
					Expect(*params.DockerImage).To(Equal("sample/dockerImage"))
				})

				It("does not upload appbits", func() {
					callPush("testApp", "--docker-image", "sample/dockerImage")

					Expect(actor.UploadAppCallCount()).To(Equal(0))
					Expect(ui.Outputs).ToNot(ContainSubstrings(
						[]string{"Uploading testApp"},
					))
				})
			})

			Context("when health-check-type '-u' or '--health-check-type' is supplied", func() {
				It("shows error if value is not 'port' or none'", func() {
					callPush("app-name", "-u", "bad-value")

					Ω(ui.Outputs).To(ContainSubstrings([]string{"Error", "Invalid health-check-type", "bad-value"}))
				})

				It("does not show error if value is 'port' or none'", func() {
					callPush("app-name", "--health-check-type", "port")

					Ω(ui.Outputs).ToNot(ContainSubstrings([]string{"Error", "Invalid health-check-type", "bad-value"}))
				})
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

					domainRepo.ListDomainsForOrgStub = func(orgGuid string, cb func(models.DomainFields) bool) error {
						cb(privateDomain)
						cb(sharedDomain)
						return nil
					}

					callPush("-t", "111", "--route-path", "the-route-path", "app-name")

					Expect(routeRepo.FindCallCount()).To(Equal(1))
					host, _, _ := routeRepo.FindArgsForCall(0)
					Expect(host).To(Equal("app-name"))

					Expect(routeRepo.CreateCallCount()).To(Equal(1))
					createdHost, createdDomainFields, createdPath := routeRepo.CreateArgsForCall(0)
					Expect(createdHost).To(Equal("app-name"))
					Expect(createdDomainFields.Guid).To(Equal("shared-domain-guid"))
					Expect(createdPath).To(Equal("the-route-path"))

					Expect(routeRepo.BindCallCount()).To(Equal(1))
					boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
					Expect(boundAppGUID).To(Equal("app-name-guid"))
					Expect(boundRouteGUID).To(Equal("app-name-route-guid"))

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating app", "app-name", "my-org", "my-space"},
						[]string{"OK"},
						[]string{"Creating", "app-name.shared.cf-app.com/the-route-path"},
						[]string{"OK"},
						[]string{"Binding", "app-name.shared.cf-app.com/the-route-path"},
						[]string{"OK"},
						[]string{"Uploading app-name"},
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

					domainRepo.ListDomainsForOrgStub = func(orgGuid string, cb func(models.DomainFields) bool) error {
						cb(privateDomain)
						return nil
					}

					appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))

					callPush("-t", "111", "app-name")

					Expect(routeRepo.FindCallCount()).To(Equal(1))
					host, _, _ := routeRepo.FindArgsForCall(0)
					Expect(host).To(Equal("app-name"))

					Expect(routeRepo.CreateCallCount()).To(Equal(1))
					createdHost, createdDomainFields, createdPath := routeRepo.CreateArgsForCall(0)
					Expect(createdHost).To(Equal("app-name"))
					Expect(createdDomainFields.Guid).To(Equal("private-domain-guid"))
					Expect(createdPath).To(BeEmpty())

					Expect(routeRepo.BindCallCount()).To(Equal(1))
					boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
					Expect(boundAppGUID).To(Equal("app-name-guid"))
					Expect(boundRouteGUID).To(Equal("app-name-route-guid"))

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Creating app", "app-name", "my-org", "my-space"},
						[]string{"OK"},
						[]string{"Creating", "app-name.private.cf-app.com"},
						[]string{"OK"},
						[]string{"Binding", "app-name.private.cf-app.com"},
						[]string{"OK"},
						[]string{"Uploading app-name"},
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
					callPush("--random-route", "app-name")
					Expect(routeRepo.FindCallCount()).To(Equal(1))
					host, _, _ := routeRepo.FindArgsForCall(0)
					Expect(host).To(Equal("app-name-random-host"))
				})

				It("provides a random hostname when the random-route option is set in the manifest", func() {
					manifestApp.Set("random-route", true)

					callPush("app-name")

					Expect(routeRepo.FindCallCount()).To(Equal(1))
					host, _, _ := routeRepo.FindArgsForCall(0)
					Expect(host).To(Equal("app-name-random-host"))
				})
			})

			It("includes the app files in dir", func() {
				expectedLocalFiles := []models.AppFileFields{
					{
						Path: "the-path",
					},
					{
						Path: "the-other-path",
					},
				}
				appfiles.AppFilesInDirReturns(expectedLocalFiles, nil)
				callPush("-p", "../some/path-to/an-app/file.zip", "app-with-path")

				actualLocalFiles, _, _ := actor.GatherFilesArgsForCall(0)
				Expect(actualLocalFiles).To(Equal(expectedLocalFiles))
			})

			It("prints a message when there are no app files to process", func() {
				appfiles.AppFilesInDirReturns([]models.AppFileFields{}, nil)
				callPush("-p", "../some/path-to/an-app/file.zip", "app-with-path")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"No app files found in '../some/path-to/an-app/file.zip'"},
				))
			})

			It("prints a message when there is an error getting app files", func() {
				appfiles.AppFilesInDirReturns([]models.AppFileFields{}, errors.New("some error"))
				callPush("-p", "../some/path-to/an-app/file.zip", "app-with-path")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Error processing app files in '../some/path-to/an-app/file.zip': some error"},
				))
			})

			It("pushes the contents of the app directory or zip file specified using the -p flag", func() {
				callPush("-p", "../some/path-to/an-app/file.zip", "app-with-path")

				_, appDir, _ := actor.GatherFilesArgsForCall(0)
				Expect(appDir).To(Equal("../some/path-to/an-app/file.zip"))
			})

			It("pushes the contents of the current working directory by default", func() {
				callPush("app-with-default-path")
				dir, _ := os.Getwd()

				_, appDir, _ := actor.GatherFilesArgsForCall(0)
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
				params := appRepo.CreateArgsForCall(0)
				Expect(*params.Name).To(Equal("app-name"))
			})

			It("pushes an app when provided a manifest with one app defined", func() {
				domainRepo.FindByNameInOrgReturns(models.DomainFields{
					Name: "manifest-example.com",
					Guid: "bar-domain-guid",
				}, nil)

				manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()

				callPush()

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "manifest-host.manifest-example.com"},
					[]string{"OK"},
					[]string{"Binding", "manifest-host.manifest-example.com"},
					[]string{"manifest-app-name"},
				))

				params := appRepo.CreateArgsForCall(0)
				Expect(*params.Name).To(Equal("manifest-app-name"))
				Expect(*params.Memory).To(Equal(int64(128)))
				Expect(*params.InstanceCount).To(Equal(1))
				Expect(*params.StackName).To(Equal("custom-stack"))
				Expect(*params.BuildpackUrl).To(Equal("some-buildpack"))
				Expect(*params.Command).To(Equal("JAVA_HOME=$PWD/.openjdk JAVA_OPTS=\"-Xss995K\" ./bin/start.sh run"))
				// Expect(actor.UploadedDir).To(Equal(filepath.Clean("some/path/from/manifest"))) TODO: Re-enable this once we develop a strategy

				Expect(*params.EnvironmentVars).To(Equal(map[string]interface{}{
					"PATH": "/u/apps/my-app/bin",
					"FOO":  "baz",
				}))
			})

			It("pushes an app with multiple routes when multiple hosts are provided", func() {
				domainRepo.FindByNameInOrgReturns(models.DomainFields{
					Name: "manifest-example.com",
					Guid: "bar-domain-guid",
				}, nil)

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
				domainRepo.FindByNameInOrgReturns(models.DomainFields{
					Name: "bar.cf-app.com",
					Guid: "bar-domain-guid",
				}, nil)

				callPush("--no-route", "app-name")

				params := appRepo.CreateArgsForCall(0)
				Expect(*params.Name).To(Equal("app-name"))
				Expect(routeRepo.CreateCallCount()).To(BeZero())
			})

			It("maps the root domain route to the app when given the --no-hostname flag", func() {
				domainRepo.ListDomainsForOrgStub = func(orgGuid string, cb func(models.DomainFields) bool) error {
					cb(models.DomainFields{
						Name:   "bar.cf-app.com",
						Guid:   "bar-domain-guid",
						Shared: true,
					})

					return nil
				}

				routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "uh oh"))

				callPush("--no-hostname", "app-name")

				params := appRepo.CreateArgsForCall(0)
				Expect(*params.Name).To(Equal("app-name"))
				Expect(routeRepo.CreateCallCount()).To(Equal(1))
				createdHost, createdDomainFields, _ := routeRepo.CreateArgsForCall(0)
				Expect(createdHost).To(Equal(""))
				Expect(createdDomainFields.Guid).To(Equal("bar-domain-guid"))
			})

			It("Does not create a route when the no-route property is in the manifest", func() {
				workerManifest := singleAppManifest()
				workerManifest.Data.Get("applications").([]interface{})[0].(generic.Map).Set("no-route", true)
				manifestRepo.ReadManifestReturns.Manifest = workerManifest

				callPush("app-name")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"app-name", "is a worker", "skipping route creation"}))
				Expect(routeRepo.BindCallCount()).To(BeZero())
			})

			It("fails when given an invalid memory limit", func() {
				callPush("-m", "abcM", "app-name")

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
					Expect(appRepo.CreateCallCount()).To(Equal(2))

					firstApp := appRepo.CreateArgsForCall(0)
					secondApp := appRepo.CreateArgsForCall(1)
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
					Expect(appRepo.CreateCallCount()).To(Equal(1))
					params := appRepo.CreateArgsForCall(0)
					Expect(*params.Name).To(Equal("app2"))
				})

				It("fails when given the name of an app that is not in the manifest", func() {
					callPush("non-existant-app")

					Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
					Expect(appRepo.CreateCallCount()).To(BeZero())
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

			appRepo.ReadReturns(existingApp, nil)
			appRepo.UpdateReturns(existingApp, nil)
		})

		It("resets the app's buildpack when the -b flag is provided as 'default'", func() {
			callPush("-b", "default", "existing-app")
			_, params := appRepo.UpdateArgsForCall(0)
			Expect(*params.BuildpackUrl).To(Equal(""))
		})

		It("resets the app's command when the -c flag is provided as 'default'", func() {
			callPush("-c", "default", "existing-app")
			_, params := appRepo.UpdateArgsForCall(0)
			Expect(*params.Command).To(Equal(""))
		})

		It("resets the app's buildpack when the -b flag is provided as 'null'", func() {
			callPush("-b", "null", "existing-app")
			_, params := appRepo.UpdateArgsForCall(0)
			Expect(*params.BuildpackUrl).To(Equal(""))
		})

		It("resets the app's command when the -c flag is provided as 'null'", func() {
			callPush("-c", "null", "existing-app")
			_, params := appRepo.UpdateArgsForCall(0)
			Expect(*params.Command).To(Equal(""))
		})

		It("merges env vars from the manifest with those from the server", func() {
			manifestRepo.ReadManifestReturns.Manifest = singleAppManifest()

			callPush("existing-app")

			_, params := appRepo.UpdateArgsForCall(0)
			updatedAppEnvVars := *params.EnvironmentVars
			Expect(updatedAppEnvVars["crazy"]).To(Equal("pants"))
			Expect(updatedAppEnvVars["FOO"]).To(Equal("baz"))
			Expect(updatedAppEnvVars["foo"]).To(Equal("manchu"))
			Expect(updatedAppEnvVars["PATH"]).To(Equal("/u/apps/my-app/bin"))
		})

		It("stops the app, achieving a full-downtime deploy!", func() {
			appRepo.UpdateReturns(existingApp, nil)

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
			appRepo.ReadReturns(existingApp, nil)
			appRepo.UpdateReturns(existingApp, nil)

			callPush("existing-app")

			Expect(stopper.ApplicationStopCallCount()).To(Equal(0))
		})

		It("updates the app", func() {
			existingRoute := models.RouteSummary{}
			existingRoute.Host = "existing-app"

			existingApp.Routes = []models.RouteSummary{existingRoute}
			appRepo.ReadReturns(existingApp, nil)
			appRepo.UpdateReturns(existingApp, nil)

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

			appGUID, params := appRepo.UpdateArgsForCall(0)
			Expect(appGUID).To(Equal(existingApp.Guid))
			Expect(*params.Command).To(Equal("different start command"))
			Expect(*params.InstanceCount).To(Equal(10))
			Expect(*params.Memory).To(Equal(int64(1024)))
			Expect(*params.BuildpackUrl).To(Equal("https://github.com/heroku/heroku-buildpack-different.git"))
			Expect(*params.StackGuid).To(Equal("differentStack-guid"))
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

				domainRepo.ListDomainsForOrgStub = func(orgGuid string, cb func(models.DomainFields) bool) error {
					cb(domain)
					return nil
				}
				routeRepo.FindReturns(models.Route{
					Host:   "existing-app",
					Domain: domain,
				}, nil)

				existingApp.Routes = []models.RouteSummary{models.RouteSummary{
					Guid:   "existing-route-guid",
					Host:   "existing-app",
					Domain: domain,
				}}

				appRepo.ReadReturns(existingApp, nil)
				appRepo.UpdateReturns(existingApp, nil)
			})

			It("uses the existing route when an app already has it bound", func() {
				callPush("-d", "example.com", "existing-app")

				Expect(routeRepo.CreateCallCount()).To(BeZero())
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Creating route"}))
				Expect(ui.Outputs).To(ContainSubstrings([]string{"Using route", "existing-app", "example.com"}))
			})

			Context("and no route-related flags are given", func() {
				Context("and there is no route in the manifest", func() {
					It("does not add a route to the app", func() {
						callPush("existing-app")

						appGuid, _, _ := actor.UploadAppArgsForCall(0)
						Expect(appGuid).To(Equal("existing-app-guid"))
						Expect(domainRepo.FindByNameInOrgCallCount()).To(BeZero())
						Expect(routeRepo.FindCallCount()).To(BeZero())
						Expect(routeRepo.CreateCallCount()).To(BeZero())
					})
				})

				Context("and there is a route in the manifest", func() {
					BeforeEach(func() {
						manifestRepo.ReadManifestReturns.Manifest = existingAppManifest()
						routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "an-error"))
						domainRepo.FindByNameInOrgReturns(models.DomainFields{Name: "example.com", Guid: "example-domain-guid"}, nil)
					})

					It("adds the route", func() {
						callPush("existing-app")
						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, _, _ := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal("new-manifest-host"))
					})
				})
			})

			It("creates and binds a route when a different domain is specified", func() {
				routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "existing-app.newdomain.com"))
				domainRepo.FindByNameInOrgReturns(models.DomainFields{Guid: "domain-guid", Name: "newdomain.com"}, nil)

				callPush("-d", "newdomain.com", "existing-app")
				domainName, domainOrgGuid := domainRepo.FindByNameInOrgArgsForCall(0)
				Expect(domainName).To(Equal("newdomain.com"))
				Expect(domainOrgGuid).To(Equal("my-org-guid"))

				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, domain, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("existing-app"))
				Expect(domain.Name).To(Equal("newdomain.com"))

				Expect(routeRepo.CreateCallCount()).To(Equal(1))
				createdHost, createdDomainFields, _ := routeRepo.CreateArgsForCall(0)
				Expect(createdHost).To(Equal("existing-app"))
				Expect(createdDomainFields.Guid).To(Equal("domain-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "existing-app.newdomain.com"},
					[]string{"OK"},
					[]string{"Binding", "existing-app.newdomain.com"},
				))
			})

			It("creates and binds a route when a different hostname is specified", func() {
				routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "new-host.newdomain.com"))

				callPush("-n", "new-host", "existing-app")

				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, domain, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("new-host"))
				Expect(domain.Name).To(Equal("example.com"))

				Expect(routeRepo.CreateCallCount()).To(Equal(1))
				createdHost, createdDomainFields, _ := routeRepo.CreateArgsForCall(0)
				Expect(createdHost).To(Equal("new-host"))
				Expect(createdDomainFields.Guid).To(Equal("domain-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "new-host.example.com"},
					[]string{"OK"},
					[]string{"Binding", "new-host.example.com"},
				))
			})

			It("removes the route when the --no-route flag is given", func() {
				callPush("--no-route", "existing-app")

				appGuid, _, _ := actor.UploadAppArgsForCall(0)
				Expect(appGuid).To(Equal("existing-app-guid"))

				Expect(domainRepo.FindByNameInOrgCallCount()).To(BeZero())
				Expect(routeRepo.FindCallCount()).To(BeZero())
				Expect(routeRepo.CreateCallCount()).To(BeZero())

				Expect(routeRepo.UnbindCallCount()).To(Equal(1))
				unboundRouteGUID, unboundAppGUID := routeRepo.UnbindArgsForCall(0)
				Expect(unboundRouteGUID).To(Equal("existing-route-guid"))
				Expect(unboundAppGUID).To(Equal("existing-app-guid"))
			})

			It("binds the root domain route to an app with a pre-existing route when the --no-hostname flag is given", func() {
				routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "existing-app.example.com"))

				callPush("--no-hostname", "existing-app")

				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, domain, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal(""))
				Expect(domain.Name).To(Equal("example.com"))

				Expect(routeRepo.CreateCallCount()).To(Equal(1))
				createdHost, createdDomainFields, _ := routeRepo.CreateArgsForCall(0)
				Expect(createdHost).To(Equal(""))
				Expect(createdDomainFields.Guid).To(Equal("domain-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "example.com"},
					[]string{"OK"},
					[]string{"Binding", "example.com"},
				))
				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"existing-app.example.com"}))
			})
		})
	})

	Describe("service instances", func() {
		BeforeEach(func() {
			serviceRepo.FindInstanceByNameStub = func(name string) (models.ServiceInstance, error) {
				return maker.NewServiceInstance(name), nil
			}

			appRepo.CreateStub = func(params models.AppParams) (models.Application, error) {
				a := models.Application{}
				a.Name = *params.Name

				return a, nil
			}

			manifestRepo.ReadManifestReturns.Manifest = manifestWithServicesAndEnv()
		})

		Context("when the service is not bound", func() {
			BeforeEach(func() {
				appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))
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
				appRepo.ReadReturns(maker.NewApp(maker.Overrides{}), nil)
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
				serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.New("Error finding instance"))
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
			appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))

			callPush("-t", "FooeyTimeout", "app-name")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Invalid", "timeout", "FooeyTimeout"},
			))
		})
	})

	Describe("displaying information about files being uploaded", func() {
		It("displays information about the files being uploaded", func() {
			appfiles.CountFilesReturns(11)
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
			routeRepo.FindReturns(models.Route{
				Host:   "existing-app",
				Domain: models.DomainFields{Name: "foo.cf-app.com"},
			}, nil)
		})

		It("suggests using 'random-route' if the default route is taken", func() {
			routeRepo.BindReturns(errors.NewHttpError(400, errors.INVALID_RELATION, "The URL not available"))

			callPush("existing-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"existing-app.foo.cf-app.com", "already in use"},
				[]string{"TIP", "random-route"},
			))
		})

		It("does not suggest using 'random-route' for other failures", func() {
			routeRepo.BindReturns(errors.NewHttpError(500, "some-code", "exception happened"))

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
			[]string{"USAGE:"},
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
