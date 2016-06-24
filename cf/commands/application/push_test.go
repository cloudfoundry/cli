package application_test

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/actors/actorsfakes"
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/api/applications/applicationsfakes"
	"github.com/cloudfoundry/cli/cf/api/authentication/authenticationfakes"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/api/stacks/stacksfakes"
	"github.com/cloudfoundry/cli/cf/appfiles/appfilesfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/commands/application/applicationfakes"
	"github.com/cloudfoundry/cli/cf/commands/service/servicefakes"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/manifest/manifestfakes"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/terminal/terminalfakes"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/generic"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/words/generator/generatorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Push Command", func() {
	var (
		cmd                        application.Push
		ui                         *terminalfakes.FakeUI
		configRepo                 coreconfig.Repository
		manifestRepo               *manifestfakes.FakeRepository
		starter                    *applicationfakes.FakeStarter
		stopper                    *applicationfakes.FakeStopper
		serviceBinder              *servicefakes.OldFakeAppBinder
		appRepo                    *applicationsfakes.FakeRepository
		domainRepo                 *apifakes.FakeDomainRepository
		routeRepo                  *apifakes.FakeRouteRepository
		stackRepo                  *stacksfakes.FakeStackRepository
		serviceRepo                *apifakes.FakeServiceRepository
		wordGenerator              *generatorfakes.FakeWordGenerator
		requirementsFactory        *requirementsfakes.FakeFactory
		authRepo                   *authenticationfakes.FakeRepository
		actor                      *actorsfakes.FakePushActor
		appfiles                   *appfilesfakes.FakeAppFiles
		zipper                     *appfilesfakes.FakeZipper
		deps                       commandregistry.Dependency
		flagContext                flags.FlagContext
		loginReq                   requirements.Passing
		targetedSpaceReq           requirements.Passing
		usageReq                   requirements.Passing
		minVersionReq              requirements.Passing
		OriginalCommandStart       commandregistry.Command
		OriginalCommandStop        commandregistry.Command
		OriginalCommandServiceBind commandregistry.Command
	)

	BeforeEach(func() {
		//save original command dependences and restore later
		OriginalCommandStart = commandregistry.Commands.FindCommand("start")
		OriginalCommandStop = commandregistry.Commands.FindCommand("stop")
		OriginalCommandServiceBind = commandregistry.Commands.FindCommand("bind-service")

		requirementsFactory = new(requirementsfakes.FakeFactory)
		loginReq = requirements.Passing{Type: "login"}
		requirementsFactory.NewLoginRequirementReturns(loginReq)
		targetedSpaceReq = requirements.Passing{Type: "targeted space"}
		requirementsFactory.NewTargetedSpaceRequirementReturns(targetedSpaceReq)
		usageReq = requirements.Passing{Type: "usage"}
		requirementsFactory.NewUsageRequirementReturns(usageReq)
		minVersionReq = requirements.Passing{Type: "minVersionReq"}
		requirementsFactory.NewMinAPIVersionRequirementReturns(minVersionReq)

		ui = new(terminalfakes.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		manifestRepo = new(manifestfakes.FakeRepository)
		wordGenerator = new(generatorfakes.FakeWordGenerator)
		wordGenerator.BabbleReturns("random-host")
		actor = new(actorsfakes.FakePushActor)
		zipper = new(appfilesfakes.FakeZipper)
		appfiles = new(appfilesfakes.FakeAppFiles)

		deps = commandregistry.Dependency{
			UI:            ui,
			Config:        configRepo,
			ManifestRepo:  manifestRepo,
			WordGenerator: wordGenerator,
			PushActor:     actor,
			AppZipper:     zipper,
			AppFiles:      appfiles,
		}

		appRepo = new(applicationsfakes.FakeRepository)
		domainRepo = new(apifakes.FakeDomainRepository)
		routeRepo = new(apifakes.FakeRouteRepository)
		serviceRepo = new(apifakes.FakeServiceRepository)
		stackRepo = new(stacksfakes.FakeStackRepository)
		authRepo = new(authenticationfakes.FakeRepository)
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo)
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.RepoLocator = deps.RepoLocator.SetStackRepository(stackRepo)
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)

		//setup fake commands (counterfeiter) to correctly interact with commandregistry
		starter = new(applicationfakes.FakeStarter)
		starter.SetDependencyStub = func(_ commandregistry.Dependency, _ bool) commandregistry.Command {
			return starter
		}
		starter.MetaDataReturns(commandregistry.CommandMetadata{Name: "start"})
		commandregistry.Register(starter)

		stopper = new(applicationfakes.FakeStopper)
		stopper.SetDependencyStub = func(_ commandregistry.Dependency, _ bool) commandregistry.Command {
			return stopper
		}
		stopper.MetaDataReturns(commandregistry.CommandMetadata{Name: "stop"})
		commandregistry.Register(stopper)

		//inject fake commands dependencies into registry
		serviceBinder = new(servicefakes.OldFakeAppBinder)
		commandregistry.Register(serviceBinder)

		cmd = application.Push{}
		cmd.SetDependency(deps, false)
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
	})

	AfterEach(func() {
		commandregistry.Register(OriginalCommandStart)
		commandregistry.Register(OriginalCommandStop)
		commandregistry.Register(OriginalCommandServiceBind)
	})

	Describe("Requirements", func() {
		var reqs []requirements.Requirement

		BeforeEach(func() {
			err := flagContext.Parse("app-name")
			Expect(err).NotTo(HaveOccurred())

			reqs = cmd.Requirements(requirementsFactory, flagContext)
		})

		It("checks that the user is logged in", func() {
			Expect(requirementsFactory.NewLoginRequirementCallCount()).To(Equal(1))
			Expect(reqs).To(ContainElement(loginReq))
		})

		It("checks that the space is targeted", func() {
			Expect(requirementsFactory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))
			Expect(reqs).To(ContainElement(targetedSpaceReq))
		})

		It("checks the number of args", func() {
			Expect(requirementsFactory.NewUsageRequirementCallCount()).To(Equal(1))
			Expect(reqs).To(ContainElement(usageReq))
		})

		Context("when --route-path is passed in", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "--route-path", "the-path")
				Expect(err).NotTo(HaveOccurred())

				reqs = cmd.Requirements(requirementsFactory, flagContext)
			})

			It("returns a minAPIVersionRequirement", func() {
				Expect(requirementsFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))

				option, version := requirementsFactory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(option).To(Equal("Option '--route-path'"))
				Expect(version).To(Equal(cf.RoutePathMinimumAPIVersion))

				Expect(reqs).To(ContainElement(minVersionReq))
			})
		})

		Context("when --app-ports is passed in", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "--app-ports", "the-app-port")
				Expect(err).NotTo(HaveOccurred())

				reqs = cmd.Requirements(requirementsFactory, flagContext)
			})

			It("returns a minAPIVersionRequirement", func() {
				Expect(requirementsFactory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))

				option, version := requirementsFactory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(option).To(Equal("Option '--app-ports'"))
				Expect(version).To(Equal(cf.MultipleAppPortsMinimumAPIVersion))

				Expect(reqs).To(ContainElement(minVersionReq))
			})
		})
	})

	Describe("Execute", func() {
		var (
			executeErr     error
			args           []string
			uiWithContents terminal.UI
			output         *gbytes.Buffer
		)

		BeforeEach(func() {
			output = gbytes.NewBuffer()
			uiWithContents = terminal.NewUI(gbytes.NewBuffer(), output, terminal.NewTeePrinter(output), trace.NewWriterPrinter(output, false))

			domainRepo.FirstOrDefaultStub = func(orgGUID string, name *string) (models.DomainFields, error) {
				if name == nil {
					var foundDomain *models.DomainFields
					domainRepo.ListDomainsForOrg(orgGUID, func(domain models.DomainFields) bool {
						foundDomain = &domain
						return !domain.Shared
					})

					if foundDomain == nil {
						return models.DomainFields{}, errors.New("Could not find a default domain")
					}

					return *foundDomain, nil
				}

				return domainRepo.FindByNameInOrg(*name, orgGUID)
			}

			domainRepo.ListDomainsForOrgStub = func(orgGUID string, cb func(models.DomainFields) bool) error {
				cb(models.DomainFields{
					Name:   "foo.cf-app.com",
					GUID:   "foo-domain-guid",
					Shared: true,
				})
				return nil
			}

			actor.ProcessPathStub = func(dirOrZipFile string, cb func(string) error) error {
				return cb(dirOrZipFile)
			}

			actor.ValidateAppParamsReturns(nil)

			appfiles.AppFilesInDirReturns(
				[]models.AppFileFields{
					{
						Path: "some-path",
					},
				},
				nil,
			)

			zipper.ZipReturns(nil)
			zipper.GetZipSizeReturns(9001, nil)
		})

		AfterEach(func() {
			output.Close()
		})

		JustBeforeEach(func() {
			cmd.SetDependency(deps, false)

			err := flagContext.Parse(args...)
			Expect(err).NotTo(HaveOccurred())

			executeErr = cmd.Execute(flagContext)
		})

		Context("when pushing a new app", func() {
			BeforeEach(func() {
				m := &manifest.Manifest{
					Path: "manifest.yml",
					Data: generic.NewMap(map[interface{}]interface{}{
						"applications": []interface{}{
							generic.NewMap(map[interface{}]interface{}{
								"name":      "manifest-app-name",
								"memory":    "128MB",
								"instances": 1,
								"host":      "manifest-host",
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
				manifestRepo.ReadManifestReturns(m, nil)

				appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))
				appRepo.CreateStub = func(params models.AppParams) (models.Application, error) {
					a := models.Application{}
					a.GUID = *params.Name + "-guid"
					a.Name = *params.Name
					a.State = "stopped"

					return a, nil
				}

				args = []string{"app-name"}
			})

			Context("validating a manifest", func() {
				BeforeEach(func() {
					actor.ValidateAppParamsReturns([]error{
						errors.New("error1"),
						errors.New("error2"),
					})
				})

				It("returns an properly formatted error", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr.Error()).To(MatchRegexp("invalid application configuration:\nerror1\nerror2"))
				})
			})

			It("tries to find the default route for the app", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(routeRepo.FindCallCount()).To(Equal(1))

				host, domain, path, port := routeRepo.FindArgsForCall(0)

				Expect(host).To(Equal("manifest-host"))
				Expect(domain.Name).To(Equal("foo.cf-app.com"))
				Expect(path).To(Equal(""))
				Expect(port).To(Equal(0))
			})

			Context("when given a bad path", func() {
				BeforeEach(func() {
					actor.ProcessPathStub = func(dirOrZipFile string, f func(string) error) error {
						return errors.New("process-path-error")
					}
				})

				JustBeforeEach(func() {
					err := flagContext.Parse("app-name", "-p", "badpath")
					Expect(err).NotTo(HaveOccurred())

					executeErr = cmd.Execute(flagContext)
				})

				It("fails with bad path error", func() {
					Expect(executeErr).To(HaveOccurred())

					Expect(executeErr.Error()).To(ContainSubstring("Error processing app files: process-path-error"))
				})
			})

			Context("when the default route for the app already exists", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(
						models.Route{
							GUID: "my-route-guid",
							Host: "app-name",
							Domain: models.DomainFields{
								Name:   "foo.cf-app.com",
								GUID:   "foo-domain-guid",
								Shared: true,
							},
						},
						nil,
					)
				})

				Context("when binding the app", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
					})

					It("binds to existing routes", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeRepo.CreateCallCount()).To(BeZero())

						Expect(routeRepo.FindCallCount()).To(Equal(1))
						host, _, _, _ := routeRepo.FindArgsForCall(0)
						Expect(host).To(Equal("manifest-host"))

						Expect(routeRepo.BindCallCount()).To(Equal(1))
						boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
						Expect(boundAppGUID).To(Equal("app-name-guid"))
						Expect(boundRouteGUID).To(Equal("my-route-guid"))

						fullOutput := terminal.Decolorize(string(output.Contents()))
						Expect(fullOutput).To(ContainSubstring("Using route app-name.foo.cf-app.com"))
						Expect(fullOutput).To(ContainSubstring("Binding app-name.foo.cf-app.com to app-name...\nOK"))
					})
				})

				Context("when pushing the app", func() {
					BeforeEach(func() {
						actor.GatherFilesReturns([]resources.AppFileResource{}, false, errors.New("failed to get file mode"))
					})

					It("notifies users about the error actor.GatherFiles() returns", func() {
						Expect(executeErr).To(HaveOccurred())

						Expect(executeErr.Error()).To(ContainSubstring("failed to get file mode"))
					})
				})
			})

			Context("when the default route for the app does not exist", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "couldn't find it"))
				})

				It("refreshes the auth token (so fresh)", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(authRepo.RefreshAuthTokenCallCount()).To(Equal(1))
				})

				Context("when refreshing the auth token fails", func() {
					BeforeEach(func() {
						authRepo.RefreshAuthTokenReturns("", errors.New("I accidentally the UAA"))
					})

					It("it returns an error", func() {
						Expect(executeErr).To(HaveOccurred())

						Expect(executeErr.Error()).To(Equal("I accidentally the UAA"))
					})
				})

				Context("when multiple domains are specified in manifest", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						domainRepo.FindByNameInOrgStub = func(name string, owningOrgGUID string) (models.DomainFields, error) {
							return map[string]models.DomainFields{
								"example1.com": {Name: "example1.com", GUID: "example-domain-guid"},
								"example2.com": {Name: "example2.com", GUID: "example-domain-guid"},
							}[name], nil
						}

						m := &manifest.Manifest{
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
						manifestRepo.ReadManifestReturns(m, nil)
						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}
						args = []string{}
					})

					It("creates a route for each domain", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						totalOutput := terminal.Decolorize(string(output.Contents()))

						Expect(totalOutput).To(ContainSubstring("Creating app manifest-app-name in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Binding manifest-host.example1.com to manifest-app-name...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Creating route manifest-host.example2.com...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Binding manifest-host.example2.com to manifest-app-name...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Creating route host2.example1.com...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Binding host2.example1.com to manifest-app-name...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Creating route host2.example2.com...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Binding host2.example2.com to manifest-app-name...\nOK"))
					})

					Context("when overriding the manifest with flags", func() {
						BeforeEach(func() {
							args = []string{"-d", "example1.com"}
						})

						It("`-d` from argument will override the domains", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							totalOutput := terminal.Decolorize(string(output.Contents()))
							Expect(totalOutput).To(ContainSubstring("Creating route manifest-host.example1.com...\nOK"))
							Expect(totalOutput).To(ContainSubstring("Binding manifest-host.example1.com"))

							Expect(totalOutput).NotTo(ContainSubstring("Creating route manifest-host.example2.com"))
						})
					})

				})

				Context("when pushing an app", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}
						args = []string{"-t", "111", "app-name"}
					})

					It("doesn't error", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						totalOutput := terminal.Decolorize(string(output.Contents()))

						Expect(totalOutput).NotTo(ContainSubstring("FAILED"))

						params := appRepo.CreateArgsForCall(0)
						Expect(*params.Name).To(Equal("app-name"))
						Expect(*params.SpaceGUID).To(Equal("my-space-guid"))

						Expect(routeRepo.FindCallCount()).To(Equal(1))
						host, _, _, _ := routeRepo.FindArgsForCall(0)
						Expect(host).To(Equal("manifest-host"))

						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, createdDomainFields, createdPath, randomPort := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal("manifest-host"))
						Expect(createdDomainFields.GUID).To(Equal("foo-domain-guid"))
						Expect(createdPath).To(BeEmpty())
						Expect(randomPort).To(BeFalse())

						Expect(routeRepo.BindCallCount()).To(Equal(1))
						boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
						Expect(boundAppGUID).To(Equal("app-name-guid"))
						Expect(boundRouteGUID).To(Equal("my-route-guid"))

						Expect(actor.UploadAppCallCount()).To(Equal(1))
						appGUID, _, _ := actor.UploadAppArgsForCall(0)
						Expect(appGUID).To(Equal("app-name-guid"))

						Expect(totalOutput).To(ContainSubstring("Creating app app-name in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Creating route manifest-host.foo.cf-app.com...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Binding manifest-host.foo.cf-app.com to app-name...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Uploading app-name...\nOK"))

						Expect(stopper.ApplicationStopCallCount()).To(Equal(0))

						app, orgName, spaceName := starter.ApplicationStartArgsForCall(0)
						Expect(app.GUID).To(Equal(appGUID))
						Expect(app.Name).To(Equal("app-name"))
						Expect(orgName).To(Equal(configRepo.OrganizationFields().Name))
						Expect(spaceName).To(Equal(configRepo.SpaceFields().Name))
						Expect(starter.SetStartTimeoutInSecondsArgsForCall(0)).To(Equal(111))
					})
				})

				Context("when there are special characters in the app name", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}
						args = []string{"-t", "111", "app!name"}
					})

					It("strips special characters when creating a default route", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeRepo.FindCallCount()).To(Equal(1))
						host, _, _, _ := routeRepo.FindArgsForCall(0)
						Expect(host).To(Equal("manifest-host"))
						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, _, _, _ := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal("manifest-host"))

						totalOutputs := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutputs).To(ContainSubstring("Creating route manifest-host.foo.cf-app.com"))
						Expect(totalOutputs).To(ContainSubstring("Binding manifest-host.foo.cf-app.com"))
						Expect(totalOutputs).NotTo(ContainSubstring("FAILED"))
					})
				})

				Context("when flags are provided", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}
						domainRepo.FindByNameInOrgReturns(models.DomainFields{
							Name: "bar.cf-app.com",
							GUID: "bar-domain-guid",
						}, nil)
						stackRepo.FindByNameReturns(models.Stack{
							Name: "customLinux",
							GUID: "custom-linux-guid",
						}, nil)

						args = []string{
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
							"--app-ports", "8080,9000",
							"--no-start",
							"app-name",
						}
					})

					It("sets the app params from the flags", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						totalOutput := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutput).To(ContainSubstring("Using stack customLinux...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Creating app app-name in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Creating route my-hostname.bar.cf-app.com/my-route-path...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Binding my-hostname.bar.cf-app.com to app-name"))
						Expect(totalOutput).To(ContainSubstring("Uploading app-name...\nOK"))

						Expect(stackRepo.FindByNameArgsForCall(0)).To(Equal("customLinux"))

						params := appRepo.CreateArgsForCall(0)
						Expect(*params.Name).To(Equal("app-name"))
						Expect(*params.Command).To(Equal("unicorn -c config/unicorn.rb -D"))
						Expect(*params.InstanceCount).To(Equal(3))
						Expect(*params.DiskQuota).To(Equal(int64(4096)))
						Expect(*params.Memory).To(Equal(int64(2048)))
						Expect(*params.StackGUID).To(Equal("custom-linux-guid"))
						Expect(*params.HealthCheckTimeout).To(Equal(1))
						Expect(*params.HealthCheckType).To(Equal("port"))
						Expect(*params.BuildpackURL).To(Equal("https://github.com/heroku/heroku-buildpack-play.git"))
						Expect(*params.AppPorts).To(Equal([]int{8080, 9000}))

						name, owningOrgGUID := domainRepo.FindByNameInOrgArgsForCall(0)
						Expect(name).To(Equal("bar.cf-app.com"))
						Expect(owningOrgGUID).To(Equal("my-org-guid"))

						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, createdDomainFields, createdPath, randomPort := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal("my-hostname"))
						Expect(createdDomainFields.GUID).To(Equal("bar-domain-guid"))
						Expect(createdPath).To(Equal("my-route-path"))
						Expect(randomPort).To(BeFalse())

						Expect(routeRepo.BindCallCount()).To(Equal(1))
						boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
						Expect(boundAppGUID).To(Equal("app-name-guid"))
						Expect(boundRouteGUID).To(Equal("my-route-guid"))

						appGUID, _, _ := actor.UploadAppArgsForCall(0)
						Expect(appGUID).To(Equal("app-name-guid"))

						Expect(starter.ApplicationStartCallCount()).To(Equal(0))
					})
				})

				Context("when an invalid app port is porvided", func() {
					BeforeEach(func() {
						args = []string{"--app-ports", "8080abc", "app-name"}
					})

					It("returns an error", func() {
						Expect(executeErr).To(HaveOccurred())

						Expect(executeErr.Error()).To(ContainSubstring("Invalid app port: 8080abc"))
						Expect(executeErr.Error()).To(ContainSubstring("App port must be a number"))
					})
				})

				Context("when pushing a docker image with --docker-image", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						args = []string{"testApp", "--docker-image", "sample/dockerImage"}
					})

					It("sets diego to true", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(appRepo.CreateCallCount()).To(Equal(1))
						params := appRepo.CreateArgsForCall(0)
						Expect(*params.Diego).To(BeTrue())
					})

					It("sets docker_image", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						params := appRepo.CreateArgsForCall(0)
						Expect(*params.DockerImage).To(Equal("sample/dockerImage"))
					})

					It("does not upload appbits", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(actor.UploadAppCallCount()).To(Equal(0))
						Expect(terminal.Decolorize(string(output.Contents()))).NotTo(ContainSubstring("Uploading testApp"))
					})

					Context("when using -o alias", func() {
						BeforeEach(func() {
							args = []string{"testApp", "-o", "sample/dockerImage"}
						})

						It("sets docker_image", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							params := appRepo.CreateArgsForCall(0)
							Expect(*params.DockerImage).To(Equal("sample/dockerImage"))
						})
					})
				})

				Context("when health-check-type '-u' or '--health-check-type' is set", func() {
					Context("when the value is not 'port' or 'none'", func() {
						BeforeEach(func() {
							args = []string{"app-name", "-u", "bad-value"}
						})

						It("returns an error", func() {
							Expect(executeErr).To(HaveOccurred())

							Expect(executeErr.Error()).To(ContainSubstring("Error: Invalid health-check-type param: bad-value"))
						})
					})

					Context("when the value is 'port'", func() {
						BeforeEach(func() {
							args = []string{"app-name", "--health-check-type", "port"}
						})

						It("does not show error", func() {
							Expect(executeErr).NotTo(HaveOccurred())
						})
					})

					Context("when the value is 'none'", func() {
						BeforeEach(func() {
							args = []string{"app-name", "--health-check-type", "none"}
						})

						It("does not show error", func() {
							Expect(executeErr).NotTo(HaveOccurred())
						})
					})
				})

				Context("when there is a shared domain", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}
						privateDomain := models.DomainFields{
							Shared: false,
							Name:   "private.cf-app.com",
							GUID:   "private-domain-guid",
						}
						sharedDomain := models.DomainFields{
							Name:   "shared.cf-app.com",
							Shared: true,
							GUID:   "shared-domain-guid",
						}

						domainRepo.ListDomainsForOrgStub = func(orgGUID string, cb func(models.DomainFields) bool) error {
							cb(privateDomain)
							cb(sharedDomain)
							return nil
						}

						args = []string{"-t", "111", "--route-path", "the-route-path", "app-name"}
					})

					It("creates a route with the shared domain and maps it to the app", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeRepo.FindCallCount()).To(Equal(1))
						host, _, _, _ := routeRepo.FindArgsForCall(0)
						Expect(host).To(Equal("manifest-host"))

						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, createdDomainFields, createdPath, randomPort := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal("manifest-host"))
						Expect(createdDomainFields.GUID).To(Equal("shared-domain-guid"))
						Expect(createdPath).To(Equal("the-route-path"))
						Expect(randomPort).To(BeFalse())

						Expect(routeRepo.BindCallCount()).To(Equal(1))
						boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
						Expect(boundAppGUID).To(Equal("app-name-guid"))
						Expect(boundRouteGUID).To(Equal("my-route-guid"))

						totalOutput := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutput).To(ContainSubstring("Creating app app-name in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Creating route manifest-host.shared.cf-app.com/the-route-path...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Binding manifest-host.shared.cf-app.com to app-name...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Uploading app-name...\nOK"))
					})
				})

				Context("when there is no shared domain but there is a private domain in the targeted org", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents

						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}

						privateDomain := models.DomainFields{
							Shared: false,
							Name:   "private.cf-app.com",
							GUID:   "private-domain-guid",
						}

						domainRepo.ListDomainsForOrgStub = func(orgGUID string, cb func(models.DomainFields) bool) error {
							cb(privateDomain)
							return nil
						}

						appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))

						args = []string{"-t", "111", "app-name"}
					})

					It("creates a route with the private domain and maps it to the app", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeRepo.FindCallCount()).To(Equal(1))
						host, _, _, _ := routeRepo.FindArgsForCall(0)
						Expect(host).To(Equal("manifest-host"))

						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, createdDomainFields, createdPath, randomPort := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal("manifest-host"))
						Expect(createdDomainFields.GUID).To(Equal("private-domain-guid"))
						Expect(createdPath).To(BeEmpty())
						Expect(randomPort).To(BeFalse())

						Expect(routeRepo.BindCallCount()).To(Equal(1))
						boundRouteGUID, boundAppGUID := routeRepo.BindArgsForCall(0)
						Expect(boundAppGUID).To(Equal("app-name-guid"))
						Expect(boundRouteGUID).To(Equal("my-route-guid"))

						totalOutput := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutput).To(ContainSubstring("Creating app app-name in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Creating route manifest-host.private.cf-app.com...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Binding manifest-host.private.cf-app.com to app-name...\nOK"))
						Expect(totalOutput).To(ContainSubstring("Uploading app-name...\nOK"))
					})
				})

				Context("with random-route option set", func() {
					var manifestApp generic.Map

					BeforeEach(func() {
						manifestApp = generic.NewMap(map[interface{}]interface{}{
							"name":      "manifest-app-name",
							"memory":    "128MB",
							"instances": 1,
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
						})
						m := &manifest.Manifest{
							Path: "manifest.yml",
							Data: generic.NewMap(map[interface{}]interface{}{
								"applications": []interface{}{manifestApp},
							}),
						}
						manifestRepo.ReadManifestReturns(m, nil)
					})

					Context("for http routes", func() {
						Context("when random hostname is set as a flag", func() {
							BeforeEach(func() {
								args = []string{"--random-route", "app-name"}
							})

							It("provides a random hostname", func() {
								Expect(executeErr).NotTo(HaveOccurred())

								Expect(routeRepo.FindCallCount()).To(Equal(1))
								host, _, _, _ := routeRepo.FindArgsForCall(0)
								Expect(host).To(Equal("app-name-random-host"))
							})
						})

						Context("when random hostname is set in the manifest", func() {
							BeforeEach(func() {
								manifestApp.Set("random-route", true)
								args = []string{"app-name"}
							})

							It("provides a random hostname", func() {
								Expect(executeErr).NotTo(HaveOccurred())

								Expect(routeRepo.FindCallCount()).To(Equal(1))
								host, _, _, _ := routeRepo.FindArgsForCall(0)
								Expect(host).To(Equal("app-name-random-host"))
							})
						})
					})

					Context("for tcp routes", func() {
						var expectedDomain models.DomainFields

						BeforeEach(func() {
							deps.UI = uiWithContents

							expectedDomain = models.DomainFields{
								GUID: "some-guid",
								Name: "some-name",
								OwningOrganizationGUID: "some-organization-guid",
								RouterGroupGUID:        "some-router-group-guid",
								RouterGroupType:        "tcp",
								Shared:                 true,
							}

							domainRepo.FindByNameInOrgReturns(
								expectedDomain,
								nil,
							)

							route := models.Route{
								Domain: expectedDomain,
								Port:   7777,
							}
							routeRepo.CreateReturns(route, nil)
						})

						Context("when random-route passed as a flag", func() {
							BeforeEach(func() {
								args = []string{"--random-route", "app-name"}
							})

							It("provides a random port and hostename", func() {
								Expect(executeErr).NotTo(HaveOccurred())

								Expect(routeRepo.FindCallCount()).To(Equal(1))
								host, domain, path, port := routeRepo.FindArgsForCall(0)
								Expect(host).To(Equal(""))
								Expect(domain).To(Equal(expectedDomain))
								Expect(path).To(Equal(""))
								Expect(port).To(Equal(0))

								Expect(routeRepo.CreateCallCount()).To(Equal(1))
								host, domain, path, useRandomPort := routeRepo.CreateArgsForCall(0)
								Expect(host).To(Equal(""))
								Expect(domain).To(Equal(expectedDomain))
								Expect(path).To(Equal(""))
								Expect(useRandomPort).To(BeTrue())

								totalOutput := terminal.Decolorize(string(output.Contents()))
								Expect(totalOutput).To(ContainSubstring("Creating random route for " + expectedDomain.Name + "..."))
							})
						})

						Context("when random-route set in the manifest", func() {
							BeforeEach(func() {
								manifestApp.Set("random-route", true)
								args = []string{"app-name"}
							})

							It("provides a random port and hostname when set in the manifest", func() {
								Expect(executeErr).NotTo(HaveOccurred())

								Expect(routeRepo.FindCallCount()).To(Equal(1))
								host, domain, path, port := routeRepo.FindArgsForCall(0)
								Expect(host).To(Equal(""))
								Expect(domain).To(Equal(expectedDomain))
								Expect(path).To(Equal(""))
								Expect(port).To(Equal(0))

								Expect(routeRepo.CreateCallCount()).To(Equal(1))
								host, domain, path, useRandomPort := routeRepo.CreateArgsForCall(0)
								Expect(host).To(Equal(""))
								Expect(domain).To(Equal(expectedDomain))
								Expect(path).To(Equal(""))
								Expect(useRandomPort).To(BeTrue())

								totalOutput := terminal.Decolorize(string(output.Contents()))
								Expect(totalOutput).To(ContainSubstring("Creating random route for " + expectedDomain.Name + "..."))
							})
						})
					})
				})

				Context("when path to an app is set", func() {
					var expectedLocalFiles []models.AppFileFields

					BeforeEach(func() {
						expectedLocalFiles = []models.AppFileFields{
							{
								Path: "the-path",
							},
							{
								Path: "the-other-path",
							},
						}
						appfiles.AppFilesInDirReturns(expectedLocalFiles, nil)
						args = []string{"-p", "../some/path-to/an-app/file.zip", "app-with-path"}
					})

					It("includes the app files in dir", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						actualLocalFiles, _, _ := actor.GatherFilesArgsForCall(0)
						Expect(actualLocalFiles).To(Equal(expectedLocalFiles))
					})
				})

				Context("when there are no app files to process", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						appfiles.AppFilesInDirReturns([]models.AppFileFields{}, nil)
						args = []string{"-p", "../some/path-to/an-app/file.zip", "app-with-path"}
					})

					It("errors", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(ContainSubstring("No app files found in '../some/path-to/an-app/file.zip'"))
					})
				})

				Context("when there is an error getting app files", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						appfiles.AppFilesInDirReturns([]models.AppFileFields{}, errors.New("some error"))
						args = []string{"-p", "../some/path-to/an-app/file.zip", "app-with-path"}
					})

					It("prints a message", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(ContainSubstring("Error processing app files in '../some/path-to/an-app/file.zip': some error"))
					})
				})

				Context("when an app path is specified with the -p flag", func() {
					BeforeEach(func() {
						args = []string{"-p", "../some/path-to/an-app/file.zip", "app-with-path"}
					})

					It("pushes the contents of the app directory or zip file specified", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						_, appDir, _ := actor.GatherFilesArgsForCall(0)
						Expect(appDir).To(Equal("../some/path-to/an-app/file.zip"))
					})
				})

				Context("when no flags are specified", func() {
					BeforeEach(func() {
						m := &manifest.Manifest{
							Path: "manifest.yml",
							Data: generic.NewMap(map[interface{}]interface{}{
								"applications": []interface{}{
									generic.NewMap(map[interface{}]interface{}{
										"name":      "manifest-app-name",
										"memory":    "128MB",
										"instances": 1,
										"host":      "manifest-host",
										"stack":     "custom-stack",
										"timeout":   360,
										"buildpack": "some-buildpack",
										"command":   `JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run`,
										"env": generic.NewMap(map[interface{}]interface{}{
											"FOO":  "baz",
											"PATH": "/u/apps/my-app/bin",
										}),
									}),
								},
							}),
						}
						manifestRepo.ReadManifestReturns(m, nil)
						args = []string{"app-with-default-path"}
					})

					It("pushes the contents of the current working directory by default", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						dir, _ := os.Getwd()
						_, appDir, _ := actor.GatherFilesArgsForCall(0)
						Expect(appDir).To(Equal(dir))
					})
				})

				Context("when given a bad manifest", func() {
					BeforeEach(func() {
						manifestRepo.ReadManifestReturns(manifest.NewEmptyManifest(), errors.New("read manifest error"))
						args = []string{"-f", "bad/manifest/path"}
					})

					It("errors", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(ContainSubstring("read manifest error"))
					})
				})

				Context("when the current directory does not contain a manifest", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						manifestRepo.ReadManifestReturns(manifest.NewEmptyManifest(), syscall.ENOENT)
						args = []string{"--no-route", "app-name"}
					})

					It("does not fail", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						fullOutput := terminal.Decolorize(string(output.Contents()))
						Expect(fullOutput).To(ContainSubstring("Creating app app-name in org my-org / space my-space as my-user...\nOK"))
						Expect(fullOutput).To(ContainSubstring("Uploading app-name...\nOK"))
					})
				})

				Context("when the current directory does contain a manifest", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						m := &manifest.Manifest{
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
						manifestRepo.ReadManifestReturns(m, nil)
						args = []string{"-p", "some/relative/path"}
					})

					It("uses the manifest in the current directory by default", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(terminal.Decolorize(string(output.Contents()))).To(ContainSubstring("Using manifest file manifest.yml"))

						cwd, _ := os.Getwd()
						Expect(manifestRepo.ReadManifestArgsForCall(0)).To(Equal(cwd))
					})
				})

				Context("when the 'no-manifest'flag is passed", func() {
					BeforeEach(func() {
						args = []string{"--no-route", "--no-manifest", "app-name"}
					})

					It("does not use a manifest", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						fullOutput := terminal.Decolorize(string(output.Contents()))
						Expect(fullOutput).NotTo(ContainSubstring("FAILED"))
						Expect(fullOutput).NotTo(ContainSubstring("hacker-manifesto"))

						Expect(manifestRepo.ReadManifestCallCount()).To(BeZero())
						params := appRepo.CreateArgsForCall(0)
						Expect(*params.Name).To(Equal("app-name"))
					})
				})

				Context("when provided a manifest with one app defined", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents

						domainRepo.FindByNameInOrgReturns(models.DomainFields{
							Name: "manifest-example.com",
							GUID: "bar-domain-guid",
						}, nil)

						m := &manifest.Manifest{
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
						manifestRepo.ReadManifestReturns(m, nil)

						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}

						args = []string{}
					})

					It("pushes the app", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						fullOutput := terminal.Decolorize(string(output.Contents()))
						Expect(fullOutput).To(ContainSubstring("Creating route manifest-host.manifest-example.com...\nOK"))
						Expect(fullOutput).To(ContainSubstring("Binding manifest-host.manifest-example.com to manifest-app-name...\nOK"))
						Expect(fullOutput).To(ContainSubstring("Uploading manifest-app-name...\nOK"))

						params := appRepo.CreateArgsForCall(0)
						Expect(*params.Name).To(Equal("manifest-app-name"))
						Expect(*params.Memory).To(Equal(int64(128)))
						Expect(*params.InstanceCount).To(Equal(1))
						Expect(*params.StackName).To(Equal("custom-stack"))
						Expect(*params.BuildpackURL).To(Equal("some-buildpack"))
						Expect(*params.Command).To(Equal("JAVA_HOME=$PWD/.openjdk JAVA_OPTS=\"-Xss995K\" ./bin/start.sh run"))
						// Expect(actor.UploadedDir).To(Equal(filepath.Clean("some/path/from/manifest"))) TODO: Re-enable this once we develop a strategy

						Expect(*params.EnvironmentVars).To(Equal(map[string]interface{}{
							"PATH": "/u/apps/my-app/bin",
							"FOO":  "baz",
						}))
					})
				})

				Context("when multiple hosts are provided", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents

						domainRepo.FindByNameInOrgReturns(models.DomainFields{
							Name: "manifest-example.com",
							GUID: "bar-domain-guid",
						}, nil)

						m := &manifest.Manifest{
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
						manifestRepo.ReadManifestReturns(m, nil)

						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}

						args = []string{}
					})

					It("pushes an app with multiple routes", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						fullOutput := terminal.Decolorize(string(output.Contents()))
						Expect(fullOutput).To(ContainSubstring("Creating route manifest-host-1.manifest-example.com...\nOK"))
						Expect(fullOutput).To(ContainSubstring("Binding manifest-host-1.manifest-example.com to manifest-app-name...\nOK"))
						Expect(fullOutput).To(ContainSubstring("Creating route manifest-host-2.manifest-example.com...\nOK"))
						Expect(fullOutput).To(ContainSubstring("Binding manifest-host-2.manifest-example.com to manifest-app-name...\nOK"))
						Expect(fullOutput).To(ContainSubstring("Uploading manifest-app-name...\nOK"))
					})
				})

				Context("when the manifest has errors", func() {
					BeforeEach(func() {
						manifestRepo.ReadManifestReturns(
							&manifest.Manifest{
								Path: "/some-path/",
							},
							errors.New("buildpack should not be null"),
						)

						args = []string{}
					})

					It("fails when parsing the manifest has errors", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(ContainSubstring("Error reading manifest file:\nbuildpack should not be null"))
					})
				})

				Context("when provided the --no-route-flag", func() {
					BeforeEach(func() {
						domainRepo.FindByNameInOrgReturns(models.DomainFields{
							Name: "bar.cf-app.com",
							GUID: "bar-domain-guid",
						}, nil)

						args = []string{"--no-route", "app-name"}
					})

					It("does not create a route", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						params := appRepo.CreateArgsForCall(0)
						Expect(*params.Name).To(Equal("app-name"))
						Expect(routeRepo.CreateCallCount()).To(BeZero())
					})
				})

				Context("when provided the --no-hostname flag", func() {
					BeforeEach(func() {
						domainRepo.ListDomainsForOrgStub = func(orgGUID string, cb func(models.DomainFields) bool) error {
							cb(models.DomainFields{
								Name:   "bar.cf-app.com",
								GUID:   "bar-domain-guid",
								Shared: true,
							})

							return nil
						}
						routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "uh oh"))

						args = []string{"--no-hostname", "app-name"}
					})

					It("maps the root domain route to the app", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						params := appRepo.CreateArgsForCall(0)
						Expect(*params.Name).To(Equal("app-name"))
						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, createdDomainFields, _, _ := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal(""))
						Expect(createdDomainFields.GUID).To(Equal("bar-domain-guid"))
					})
				})

				Context("when no-route is set in the manifest", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						workerManifest := &manifest.Manifest{
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
						workerManifest.Data.Get("applications").([]interface{})[0].(generic.Map).Set("no-route", true)
						manifestRepo.ReadManifestReturns(workerManifest, nil)

						args = []string{"app-name"}
					})

					It("Does not create a route", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(terminal.Decolorize(string(output.Contents()))).To(ContainSubstring("App app-name is a worker, skipping route creation"))
						Expect(routeRepo.BindCallCount()).To(BeZero())
					})
				})

				Context("with an invalid memory limit", func() {
					BeforeEach(func() {
						args = []string{"-m", "abcM", "app-name"}
					})

					It("fails", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(ContainSubstring("Invalid memory limit: abcM"))
					})
				})

				Context("when a manifest has many apps", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						m := &manifest.Manifest{
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
						manifestRepo.ReadManifestReturns(m, nil)
						args = []string{}
					})

					It("pushes each app", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						totalOutput := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutput).To(ContainSubstring("Creating app app1"))
						Expect(totalOutput).To(ContainSubstring("Creating app app2"))
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

					Context("when a single app is given as an arg", func() {
						BeforeEach(func() {
							args = []string{"app2"}
						})

						It("pushes that single app", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							totalOutput := terminal.Decolorize(string(output.Contents()))
							Expect(totalOutput).To(ContainSubstring("Creating app app2"))
							Expect(totalOutput).ToNot(ContainSubstring("Creating app app1"))
							Expect(appRepo.CreateCallCount()).To(Equal(1))
							params := appRepo.CreateArgsForCall(0)
							Expect(*params.Name).To(Equal("app2"))
						})
					})

					Context("when the given app is not in the manifest", func() {
						BeforeEach(func() {
							args = []string{"non-existant-app"}
						})

						It("fails", func() {
							Expect(executeErr).To(HaveOccurred())
							Expect(appRepo.CreateCallCount()).To(BeZero())
						})
					})
				})
			})
		})

		Context("re-pushing an existing app", func() {
			var existingApp models.Application

			BeforeEach(func() {
				deps.UI = uiWithContents
				existingApp = models.Application{
					ApplicationFields: models.ApplicationFields{
						Name:    "existing-app",
						GUID:    "existing-app-guid",
						Command: "unicorn -c config/unicorn.rb -D",
						EnvironmentVars: map[string]interface{}{
							"crazy": "pants",
							"FOO":   "NotYoBaz",
							"foo":   "manchu",
						},
					},
				}
				manifestRepo.ReadManifestReturns(manifest.NewEmptyManifest(), nil)
				appRepo.ReadReturns(existingApp, nil)
				appRepo.UpdateReturns(existingApp, nil)
				args = []string{"existing-app"}
			})

			It("stops the app, achieving a full-downtime deploy!", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				app, orgName, spaceName := stopper.ApplicationStopArgsForCall(0)
				Expect(app.GUID).To(Equal(existingApp.GUID))
				Expect(app.Name).To(Equal("existing-app"))
				Expect(orgName).To(Equal(configRepo.OrganizationFields().Name))
				Expect(spaceName).To(Equal(configRepo.SpaceFields().Name))

				Expect(actor.UploadAppCallCount()).To(Equal(1))
				appGUID, _, _ := actor.UploadAppArgsForCall(0)
				Expect(appGUID).To(Equal(existingApp.GUID))
			})

			It("re-uploads the app", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				totalOutputs := terminal.Decolorize(string(output.Contents()))
				Expect(totalOutputs).To(ContainSubstring("Uploading existing-app...\nOK"))
			})

			Context("when the -b flag is provided as 'default'", func() {
				BeforeEach(func() {
					args = []string{"-b", "default", "existing-app"}
				})

				It("resets the app's buildpack", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(appRepo.UpdateCallCount()).To(Equal(1))
					_, params := appRepo.UpdateArgsForCall(0)
					Expect(*params.BuildpackURL).To(Equal(""))
				})
			})

			Context("when the -c flag is provided as 'default'", func() {
				BeforeEach(func() {
					args = []string{"-c", "default", "existing-app"}
				})

				It("resets the app's command", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					_, params := appRepo.UpdateArgsForCall(0)
					Expect(*params.Command).To(Equal(""))
				})
			})

			Context("when the -b flag is provided as 'null'", func() {
				BeforeEach(func() {
					args = []string{"-b", "null", "existing-app"}
				})

				It("resets the app's buildpack", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					_, params := appRepo.UpdateArgsForCall(0)
					Expect(*params.BuildpackURL).To(Equal(""))
				})
			})

			Context("when the -c flag is provided as 'null'", func() {
				BeforeEach(func() {
					args = []string{"-c", "null", "existing-app"}
				})

				It("resets the app's command", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					_, params := appRepo.UpdateArgsForCall(0)
					Expect(*params.Command).To(Equal(""))
				})
			})

			Context("when the manifest provided env variables", func() {
				BeforeEach(func() {
					m := &manifest.Manifest{
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
					manifestRepo.ReadManifestReturns(m, nil)

					args = []string{"existing-app"}
				})

				It("merges env vars from the manifest with those from the server", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					_, params := appRepo.UpdateArgsForCall(0)
					updatedAppEnvVars := *params.EnvironmentVars
					Expect(updatedAppEnvVars["crazy"]).To(Equal("pants"))
					Expect(updatedAppEnvVars["FOO"]).To(Equal("baz"))
					Expect(updatedAppEnvVars["foo"]).To(Equal("manchu"))
					Expect(updatedAppEnvVars["PATH"]).To(Equal("/u/apps/my-app/bin"))
				})
			})

			Context("when the app is already stopped", func() {
				BeforeEach(func() {
					existingApp.State = "stopped"
					appRepo.ReadReturns(existingApp, nil)
					appRepo.UpdateReturns(existingApp, nil)
					args = []string{"existing-app"}
				})

				It("does not stop the app", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(stopper.ApplicationStopCallCount()).To(Equal(0))
				})
			})

			Context("when the application is pushed with updated parameters", func() {
				BeforeEach(func() {
					existingRoute := models.RouteSummary{
						Host: "existing-app",
					}
					existingApp.Routes = []models.RouteSummary{existingRoute}
					appRepo.ReadReturns(existingApp, nil)
					appRepo.UpdateReturns(existingApp, nil)

					stackRepo.FindByNameReturns(models.Stack{
						Name: "differentStack",
						GUID: "differentStack-guid",
					}, nil)

					args = []string{
						"-c", "different start command",
						"-i", "10",
						"-m", "1G",
						"-b", "https://github.com/heroku/heroku-buildpack-different.git",
						"-s", "differentStack",
						"existing-app",
					}
				})

				It("updates the app", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					appGUID, params := appRepo.UpdateArgsForCall(0)
					Expect(appGUID).To(Equal(existingApp.GUID))
					Expect(*params.Command).To(Equal("different start command"))
					Expect(*params.InstanceCount).To(Equal(10))
					Expect(*params.Memory).To(Equal(int64(1024)))
					Expect(*params.BuildpackURL).To(Equal("https://github.com/heroku/heroku-buildpack-different.git"))
					Expect(*params.StackGUID).To(Equal("differentStack-guid"))
				})
			})

			Context("when the app has a route bound", func() {
				BeforeEach(func() {
					domain := models.DomainFields{
						Name:   "example.com",
						GUID:   "domain-guid",
						Shared: true,
					}

					domainRepo.ListDomainsForOrgStub = func(orgGUID string, cb func(models.DomainFields) bool) error {
						cb(domain)
						return nil
					}

					routeRepo.FindReturns(models.Route{
						Host:   "existing-app",
						Domain: domain,
					}, nil)

					routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ bool) (models.Route, error) {
						return models.Route{
							GUID:   "my-route-guid",
							Host:   host,
							Domain: domain,
						}, nil
					}

					existingApp.Routes = []models.RouteSummary{{
						GUID:   "existing-route-guid",
						Host:   "existing-app",
						Domain: domain,
					}}

					appRepo.ReadReturns(existingApp, nil)
					appRepo.UpdateReturns(existingApp, nil)
				})

				Context("when an existing route is set", func() {
					BeforeEach(func() {
						args = []string{"-d", "example.com", "existing-app"}
					})

					It("uses the existing route", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeRepo.CreateCallCount()).To(BeZero())
						totalOutputs := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutputs).NotTo(ContainSubstring("Creating route"))
						Expect(totalOutputs).To(ContainSubstring("Using route existing-app.example.com"))
					})
				})

				Context("and no route-related flags are given", func() {
					Context("and there is no route in the manifest", func() {
						It("does not add a route to the app", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							appGUID, _, _ := actor.UploadAppArgsForCall(0)
							Expect(appGUID).To(Equal("existing-app-guid"))
							Expect(domainRepo.FindByNameInOrgCallCount()).To(BeZero())
							Expect(routeRepo.FindCallCount()).To(BeZero())
							Expect(routeRepo.CreateCallCount()).To(BeZero())
						})
					})

					Context("and there is a route in the manifest", func() {
						BeforeEach(func() {
							m := &manifest.Manifest{
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
							manifestRepo.ReadManifestReturns(m, nil)
							routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "an-error"))
							domainRepo.FindByNameInOrgReturns(models.DomainFields{Name: "example.com", GUID: "example-domain-guid"}, nil)
						})

						It("adds the route", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(routeRepo.CreateCallCount()).To(Equal(1))
							createdHost, _, _, _ := routeRepo.CreateArgsForCall(0)
							Expect(createdHost).To(Equal("new-manifest-host"))
						})
					})
				})

				Context("when a different domain is set", func() {
					BeforeEach(func() {
						routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "existing-app.newdomain.com"))
						domainRepo.FindByNameInOrgReturns(models.DomainFields{GUID: "domain-guid", Name: "newdomain.com"}, nil)
						args = []string{"-d", "newdomain.com", "existing-app"}
					})

					It("creates and binds a route in the new domain", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(domainRepo.FirstOrDefaultCallCount()).To(Equal(1))
						domainOrgGUID, domainName := domainRepo.FirstOrDefaultArgsForCall(0)
						Expect(*domainName).To(Equal("newdomain.com"))
						Expect(domainOrgGUID).To(Equal("my-org-guid"))

						Expect(routeRepo.FindCallCount()).To(Equal(1))
						host, domain, _, _ := routeRepo.FindArgsForCall(0)
						Expect(host).To(Equal("existing-app"))
						Expect(domain.Name).To(Equal("newdomain.com"))

						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, createdDomainFields, _, randomPort := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal("existing-app"))
						Expect(createdDomainFields.GUID).To(Equal("domain-guid"))
						Expect(randomPort).To(BeFalse())

						totalOutputs := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutputs).To(ContainSubstring("Creating route existing-app.newdomain.com...\nOK"))
						Expect(totalOutputs).To(ContainSubstring("Binding existing-app.newdomain.com to existing-app...\nOK"))
					})
				})

				Context("when a different hostname is set", func() {
					BeforeEach(func() {
						domainRepo.FindByNameInOrgReturns(models.DomainFields{GUID: "domain-guid", Name: "example.com"}, nil)
						routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "new-host.newdomain.com"))
						args = []string{"-n", "new-host", "existing-app"}
					})

					It("creates and binds a route when a different hostname is specified", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeRepo.FindCallCount()).To(Equal(1))
						host, domain, _, _ := routeRepo.FindArgsForCall(0)
						Expect(host).To(Equal("new-host"))
						Expect(domain.Name).To(Equal("example.com"))

						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, createdDomainFields, _, _ := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal("new-host"))
						Expect(createdDomainFields.GUID).To(Equal("domain-guid"))

						totalOutputs := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutputs).To(ContainSubstring("Creating route new-host.example.com...\nOK"))
						Expect(totalOutputs).To(ContainSubstring("Binding new-host.example.com to existing-app...\nOK"))
					})
				})

				Context("when --no-route flag is given", func() {
					BeforeEach(func() {
						args = []string{"--no-route", "existing-app"}
					})

					It("removes existing routes that the app is bound to", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						appGUID, _, _ := actor.UploadAppArgsForCall(0)
						Expect(appGUID).To(Equal("existing-app-guid"))

						Expect(domainRepo.FindByNameInOrgCallCount()).To(BeZero())
						Expect(routeRepo.FindCallCount()).To(BeZero())
						Expect(routeRepo.CreateCallCount()).To(BeZero())

						Expect(routeRepo.UnbindCallCount()).To(Equal(1))
						unboundRouteGUID, unboundAppGUID := routeRepo.UnbindArgsForCall(0)
						Expect(unboundRouteGUID).To(Equal("existing-route-guid"))
						Expect(unboundAppGUID).To(Equal("existing-app-guid"))
					})
				})

				Context("when the --no-hostname flag is given", func() {
					BeforeEach(func() {
						routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "existing-app.example.com"))
						args = []string{"--no-hostname", "existing-app"}
					})

					It("binds the root domain route to an app with a pre-existing route", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeRepo.FindCallCount()).To(Equal(1))
						host, domain, _, _ := routeRepo.FindArgsForCall(0)
						Expect(host).To(Equal(""))
						Expect(domain.Name).To(Equal("example.com"))

						Expect(routeRepo.CreateCallCount()).To(Equal(1))
						createdHost, createdDomainFields, _, randomPort := routeRepo.CreateArgsForCall(0)
						Expect(createdHost).To(Equal(""))
						Expect(createdDomainFields.GUID).To(Equal("domain-guid"))
						Expect(randomPort).To(BeFalse())

						totalOutputs := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutputs).To(ContainSubstring("Creating route example.com...\nOK"))
						Expect(totalOutputs).To(ContainSubstring("Binding example.com to existing-app...\nOK"))
						Expect(totalOutputs).NotTo(ContainSubstring("existing-app.example.com"))
					})
				})
			})

			Context("service instances", func() {
				BeforeEach(func() {
					appRepo.CreateStub = func(params models.AppParams) (models.Application, error) {
						a := models.Application{}
						a.Name = *params.Name

						return a, nil
					}

					serviceRepo.FindInstanceByNameStub = func(name string) (models.ServiceInstance, error) {
						return models.ServiceInstance{
							ServiceInstanceFields: models.ServiceInstanceFields{Name: name},
						}, nil
					}

					m := &manifest.Manifest{
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
					manifestRepo.ReadManifestReturns(m, nil)

					args = []string{}
				})

				Context("when the service is not bound", func() {
					BeforeEach(func() {
						appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))
					})

					It("binds service instances to the app", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(len(serviceBinder.AppsToBind)).To(Equal(4))
						Expect(serviceBinder.AppsToBind[0].Name).To(Equal("app1"))
						Expect(serviceBinder.AppsToBind[1].Name).To(Equal("app1"))
						Expect(serviceBinder.InstancesToBindTo[0].Name).To(Equal("app1-service"))
						Expect(serviceBinder.InstancesToBindTo[1].Name).To(Equal("global-service"))

						Expect(serviceBinder.AppsToBind[2].Name).To(Equal("app2"))
						Expect(serviceBinder.AppsToBind[3].Name).To(Equal("app2"))
						Expect(serviceBinder.InstancesToBindTo[2].Name).To(Equal("app2-service"))
						Expect(serviceBinder.InstancesToBindTo[3].Name).To(Equal("global-service"))

						totalOutputs := terminal.Decolorize(string(output.Contents()))
						Expect(totalOutputs).To(ContainSubstring("Creating app app1 in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutputs).To(ContainSubstring("Binding service app1-service to app app1 in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutputs).To(ContainSubstring("Binding service global-service to app app1 in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutputs).To(ContainSubstring("Creating app app2 in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutputs).To(ContainSubstring("Binding service app2-service to app app2 in org my-org / space my-space as my-user...\nOK"))
						Expect(totalOutputs).To(ContainSubstring("Binding service global-service to app app2 in org my-org / space my-space as my-user...\nOK"))
					})
				})

				Context("when the app is already bound to the service", func() {
					BeforeEach(func() {
						appRepo.ReadReturns(models.Application{
							ApplicationFields: models.ApplicationFields{Name: "app-name"},
						}, nil)
						serviceBinder.BindApplicationReturns.Error = errors.NewHTTPError(500, errors.ServiceBindingAppServiceTaken, "it don't work")
					})

					It("gracefully continues", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(len(serviceBinder.AppsToBind)).To(Equal(4))
					})
				})

				Context("when the service instance can't be found", func() {
					BeforeEach(func() {
						serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.New("Error finding instance"))
					})

					It("fails with an error", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(ContainSubstring("Could not find service app1-service to bind to existing-app"))
					})
				})
			})

			Context("checking for bad flags", func() {
				BeforeEach(func() {
					appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))
					args = []string{"-t", "FooeyTimeout", "app-name"}
				})

				It("fails when a non-numeric start timeout is given", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr.Error()).To(ContainSubstring("Invalid timeout param: FooeyTimeout"))
				})
			})

			Context("displaying information about files being uploaded", func() {
				BeforeEach(func() {
					appfiles.CountFilesReturns(11)
					zipper.ZipReturns(nil)
					zipper.GetZipSizeReturns(6100000, nil)
					actor.GatherFilesReturns([]resources.AppFileResource{{Path: "path/to/app"}, {Path: "bar"}}, true, nil)
					args = []string{"appName"}
				})

				It("displays the information", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					curDir, err := os.Getwd()
					Expect(err).NotTo(HaveOccurred())

					totalOutputs := terminal.Decolorize(string(output.Contents()))
					Expect(totalOutputs).To(ContainSubstring("Uploading app files from: " + curDir))
					Expect(totalOutputs).To(ContainSubstring("Uploading 5.8M, 11 files\nOK"))
				})
			})

			Context("when the app can't be uploaded", func() {
				BeforeEach(func() {
					actor.UploadAppReturns(errors.New("Boom!"))
					args = []string{"app"}
				})

				It("fails when the app can't be uploaded", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr.Error()).To(ContainSubstring("Error uploading application"))
				})
			})

			Context("when binding the route fails", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{
						Host:   "existing-app",
						Domain: models.DomainFields{Name: "foo.cf-app.com"},
					}, nil)
					args = []string{"existing-app"}
					routeRepo.BindReturns(errors.NewHTTPError(500, "some-code", "exception happened"))
				})

				It("errors", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr.Error()).ToNot(ContainSubstring("TIP"))
				})

				Context("when the default route is taken", func() {
					BeforeEach(func() {
						routeRepo.BindReturns(errors.NewHTTPError(400, errors.InvalidRelation, "The URL not available"))
					})

					It("suggests using 'random-route' if the default route is taken", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(ContainSubstring("existing-app.foo.cf-app.com is already in use"))
						Expect(executeErr.Error()).To(ContainSubstring("TIP: Change the hostname with -n HOSTNAME or use --random-route"))
					})
				})

				Context("when no name and no manifest is given", func() {
					BeforeEach(func() {
						manifestRepo.ReadManifestReturns(manifest.NewEmptyManifest(), errors.New("No such manifest"))
						args = []string{}
					})

					It("fails", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(ContainSubstring("Manifest file is not found"))
					})
				})
			})
		})
	})
})
