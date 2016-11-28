package application_test

import (
	"os"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/actors/actorsfakes"
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/api/stacks/stacksfakes"
	"code.cloudfoundry.org/cli/cf/appfiles/appfilesfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/application"
	"code.cloudfoundry.org/cli/cf/commands/application/applicationfakes"
	"code.cloudfoundry.org/cli/cf/commands/service/servicefakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/manifest"
	"code.cloudfoundry.org/cli/cf/manifest/manifestfakes"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/util/generic"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	"code.cloudfoundry.org/cli/util/words/generator/generatorfakes"
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
		routeActor                 *actorsfakes.FakeRouteActor
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
		routeActor = new(actorsfakes.FakeRouteActor)
		zipper = new(appfilesfakes.FakeZipper)
		appfiles = new(appfilesfakes.FakeAppFiles)

		deps = commandregistry.Dependency{
			UI:            ui,
			Config:        configRepo,
			ManifestRepo:  manifestRepo,
			WordGenerator: wordGenerator,
			PushActor:     actor,
			RouteActor:    routeActor,
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

			reqs, err = cmd.Requirements(requirementsFactory, flagContext)
			Expect(err).NotTo(HaveOccurred())
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

				reqs, err = cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())
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

				reqs, err = cmd.Requirements(requirementsFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())
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
					Expect(executeErr.Error()).To(MatchRegexp("Invalid application configuration:\nerror1\nerror2"))
				})
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
				var route models.Route
				BeforeEach(func() {
					route = models.Route{
						GUID: "my-route-guid",
						Host: "app-name",
						Domain: models.DomainFields{
							Name:   "foo.cf-app.com",
							GUID:   "foo-domain-guid",
							Shared: true,
						},
					}
					routeActor.FindOrCreateRouteReturns(
						route,
						nil,
					)
				})

				Context("when binding the app", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
					})

					It("binds to existing routes", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeActor.BindRouteCallCount()).To(Equal(1))
						boundApp, boundRoute := routeActor.BindRouteArgsForCall(0)
						Expect(boundApp.GUID).To(Equal("app-name-guid"))
						Expect(boundRoute).To(Equal(route))
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
					var (
						route1 models.Route
						route2 models.Route
						route3 models.Route
						route4 models.Route
					)

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
						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ int, _ bool) (models.Route, error) {
							return models.Route{
								GUID:   "my-route-guid",
								Host:   host,
								Domain: domain,
							}, nil
						}
						args = []string{}

						route1 = models.Route{
							GUID: "route1-guid",
						}
						route2 = models.Route{
							GUID: "route2-guid",
						}
						route3 = models.Route{
							GUID: "route3-guid",
						}
						route4 = models.Route{
							GUID: "route4-guid",
						}

						callCount := 0
						routeActor.FindOrCreateRouteStub = func(hostname string, domain models.DomainFields, path string, _ int, useRandomPort bool) (models.Route, error) {
							callCount = callCount + 1
							switch callCount {
							case 1:
								Expect(hostname).To(Equal("host2"))
								Expect(domain.Name).To(Equal("example1.com"))
								Expect(path).To(BeEmpty())
								Expect(useRandomPort).To(BeFalse())
								return route1, nil
							case 2:
								Expect(hostname).To(Equal("manifest-host"))
								Expect(domain.Name).To(Equal("example1.com"))
								Expect(path).To(BeEmpty())
								Expect(useRandomPort).To(BeFalse())
								return route2, nil
							case 3:
								Expect(hostname).To(Equal("host2"))
								Expect(domain.Name).To(Equal("example2.com"))
								Expect(path).To(BeEmpty())
								Expect(useRandomPort).To(BeFalse())
								return route3, nil
							case 4:
								Expect(hostname).To(Equal("manifest-host"))
								Expect(domain.Name).To(Equal("example2.com"))
								Expect(path).To(BeEmpty())
								Expect(useRandomPort).To(BeFalse())
								return route4, nil
							default:
								Fail("should have only been called 4 times")
							}
							panic("should never have gotten this far")
						}
					})

					It("creates a route for each domain", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeActor.BindRouteCallCount()).To(Equal(4))
						app, route := routeActor.BindRouteArgsForCall(0)
						Expect(app.Name).To(Equal("manifest-app-name"))
						Expect(route).To(Equal(route1))

						app, route = routeActor.BindRouteArgsForCall(1)
						Expect(app.Name).To(Equal("manifest-app-name"))
						Expect(route).To(Equal(route2))

						app, route = routeActor.BindRouteArgsForCall(2)
						Expect(app.Name).To(Equal("manifest-app-name"))
						Expect(route).To(Equal(route3))

						app, route = routeActor.BindRouteArgsForCall(3)
						Expect(app.Name).To(Equal("manifest-app-name"))
						Expect(route).To(Equal(route4))
					})

					Context("when overriding the manifest with flags", func() {
						BeforeEach(func() {
							args = []string{"-d", "example1.com"}
							route1 = models.Route{
								GUID: "route1-guid",
							}
							route2 = models.Route{
								GUID: "route2-guid",
							}

							callCount := 0
							routeActor.FindOrCreateRouteStub = func(hostname string, domain models.DomainFields, path string, _ int, useRandomPort bool) (models.Route, error) {
								callCount = callCount + 1
								switch callCount {
								case 1:
									Expect(hostname).To(Equal("host2"))
									Expect(domain.Name).To(Equal("example1.com"))
									Expect(path).To(BeEmpty())
									Expect(useRandomPort).To(BeFalse())
									return route1, nil
								case 2:
									Expect(hostname).To(Equal("manifest-host"))
									Expect(domain.Name).To(Equal("example1.com"))
									Expect(path).To(BeEmpty())
									Expect(useRandomPort).To(BeFalse())
									return route2, nil
								default:
									Fail("should have only been called 2 times")
								}
								panic("should never have gotten this far")
							}
						})

						It("`-d` from argument will override the domains", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(routeActor.BindRouteCallCount()).To(Equal(2))
							app, route := routeActor.BindRouteArgsForCall(0)
							Expect(app.Name).To(Equal("manifest-app-name"))
							Expect(route).To(Equal(route1))

							app, route = routeActor.BindRouteArgsForCall(1)
							Expect(app.Name).To(Equal("manifest-app-name"))
							Expect(route).To(Equal(route2))
						})
					})
				})

				Context("when pushing an app", func() {
					BeforeEach(func() {
						deps.UI = uiWithContents
						routeRepo.CreateStub = func(host string, domain models.DomainFields, _ string, _ int, _ bool) (models.Route, error) {
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

						Expect(actor.UploadAppCallCount()).To(Equal(1))
						appGUID, _, _ := actor.UploadAppArgsForCall(0)
						Expect(appGUID).To(Equal("app-name-guid"))

						Expect(totalOutput).To(ContainSubstring("Creating app app-name in org my-org / space my-space as my-user...\nOK"))
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
					Context("when the app name is specified via manifest file", func() {
						BeforeEach(func() {
							m := &manifest.Manifest{
								Path: "manifest.yml",
								Data: generic.NewMap(map[interface{}]interface{}{
									"applications": []interface{}{
										generic.NewMap(map[interface{}]interface{}{
											"name":      "manifest!app-nam#",
											"memory":    "128MB",
											"instances": 1,
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

							args = []string{}
						})

						It("strips special characters when creating a default route", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(routeActor.FindOrCreateRouteCallCount()).To(Equal(1))
							host, _, _, _, _ := routeActor.FindOrCreateRouteArgsForCall(0)
							Expect(host).To(Equal("manifestapp-nam"))
						})
					})

					Context("when the app name is specified via flag", func() {
						BeforeEach(func() {
							manifestRepo.ReadManifestReturns(manifest.NewEmptyManifest(), nil)
							args = []string{"app@#name"}
						})

						It("strips special characters when creating a default route", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(routeActor.FindOrCreateRouteCallCount()).To(Equal(1))
							host, _, _, _, _ := routeActor.FindOrCreateRouteArgsForCall(0)
							Expect(host).To(Equal("appname"))
						})
					})
				})

				Context("when flags are provided", func() {
					BeforeEach(func() {
						m := &manifest.Manifest{
							Path: "manifest.yml",
							Data: generic.NewMap(map[interface{}]interface{}{
								"applications": []interface{}{
									generic.NewMap(map[interface{}]interface{}{
										"name":              "manifest!app-nam#",
										"memory":            "128MB",
										"instances":         1,
										"host":              "host-name",
										"domain":            "domain-name",
										"disk_quota":        "1G",
										"stack":             "custom-stack",
										"timeout":           360,
										"health-check-type": "none",
										"app-ports":         []interface{}{3000},
										"buildpack":         "some-buildpack",
										"command":           `JAVA_HOME=$PWD/.openjdk JAVA_OPTS="-Xss995K" ./bin/start.sh run`,
										"path":              filepath.Clean("some/path/from/manifest"),
										"env": generic.NewMap(map[interface{}]interface{}{
											"FOO":  "baz",
											"PATH": "/u/apps/my-app/bin",
										}),
									}),
								},
							}),
						}
						manifestRepo.ReadManifestReturns(m, nil)

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

						Expect(appRepo.CreateCallCount()).To(Equal(1))
						appParam := appRepo.CreateArgsForCall(0)
						Expect(*appParam.Command).To(Equal("unicorn -c config/unicorn.rb -D"))
						Expect(appParam.Domains).To(Equal([]string{"bar.cf-app.com"}))
						Expect(appParam.Hosts).To(Equal([]string{"my-hostname"}))
						Expect(*appParam.RoutePath).To(Equal("my-route-path"))
						Expect(*appParam.Name).To(Equal("app-name"))
						Expect(*appParam.InstanceCount).To(Equal(3))
						Expect(*appParam.DiskQuota).To(Equal(int64(4096)))
						Expect(*appParam.Memory).To(Equal(int64(2048)))
						Expect(*appParam.StackName).To(Equal("customLinux"))
						Expect(*appParam.HealthCheckTimeout).To(Equal(1))
						Expect(*appParam.HealthCheckType).To(Equal("port"))
						Expect(*appParam.BuildpackURL).To(Equal("https://github.com/heroku/heroku-buildpack-play.git"))
						Expect(*appParam.AppPorts).To(Equal([]int{8080, 9000}))
						Expect(*appParam.HealthCheckTimeout).To(Equal(1))
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
						Context("when random-route is set as a flag", func() {
							BeforeEach(func() {
								args = []string{"--random-route", "app-name"}
							})

							It("provides a random hostname", func() {
								Expect(executeErr).NotTo(HaveOccurred())

								Expect(routeActor.FindOrCreateRouteCallCount()).To(Equal(1))
								host, _, _, _, _ := routeActor.FindOrCreateRouteArgsForCall(0)
								Expect(host).To(Equal("app-name-random-host"))
							})
						})

						Context("when random-route is set in the manifest", func() {
							BeforeEach(func() {
								manifestApp.Set("random-route", true)
								args = []string{"app-name"}
							})

							It("provides a random hostname", func() {
								Expect(executeErr).NotTo(HaveOccurred())

								Expect(routeActor.FindOrCreateRouteCallCount()).To(Equal(1))
								host, _, _, _, _ := routeActor.FindOrCreateRouteArgsForCall(0)
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

							It("provides a random port and hostname", func() {
								Expect(executeErr).NotTo(HaveOccurred())

								Expect(routeActor.FindOrCreateRouteCallCount()).To(Equal(1))
								_, _, _, _, randomPort := routeActor.FindOrCreateRouteArgsForCall(0)
								Expect(randomPort).To(BeTrue())
							})
						})

						Context("when random-route set in the manifest", func() {
							BeforeEach(func() {
								manifestApp.Set("random-route", true)
								args = []string{"app-name"}
							})

							It("provides a random port and hostname when set in the manifest", func() {
								Expect(executeErr).NotTo(HaveOccurred())

								Expect(routeActor.FindOrCreateRouteCallCount()).To(Equal(1))
								_, _, _, _, randomPort := routeActor.FindOrCreateRouteArgsForCall(0)
								Expect(randomPort).To(BeTrue())
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

				Context("when the no-route option is set", func() {
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

						Expect(routeActor.FindOrCreateRouteCallCount()).To(Equal(1))
						_, domain, _, _, _ := routeActor.FindOrCreateRouteArgsForCall(0)
						Expect(domain.GUID).To(Equal("bar-domain-guid"))
					})

					Context("when using 'routes' in the manifest", func() {
						BeforeEach(func() {
							m := &manifest.Manifest{
								Data: generic.NewMap(map[interface{}]interface{}{
									"applications": []interface{}{
										generic.NewMap(map[interface{}]interface{}{
											"name": "app1",
											"routes": []interface{}{
												map[interface{}]interface{}{"route": "app2route1.example.com"},
											},
										}),
									},
								}),
							}
							manifestRepo.ReadManifestReturns(m, nil)
						})

						It("returns an error", func() {
							Expect(executeErr).To(HaveOccurred())
							Expect(executeErr).To(MatchError("Option '--no-hostname' cannot be used with an app manifest containing the 'routes' attribute"))
						})
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

					existingApp.Routes = []models.RouteSummary{{
						GUID:   "existing-route-guid",
						Host:   "existing-app",
						Domain: domain,
					}}

					appRepo.ReadReturns(existingApp, nil)
					appRepo.UpdateReturns(existingApp, nil)
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
				})

				Context("when --no-route flag is given", func() {
					BeforeEach(func() {
						args = []string{"--no-route", "existing-app"}
					})

					It("removes existing routes that the app is bound to", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						appGUID, _, _ := actor.UploadAppArgsForCall(0)
						Expect(appGUID).To(Equal("existing-app-guid"))

						Expect(routeActor.UnbindAllCallCount()).To(Equal(1))
						app := routeActor.UnbindAllArgsForCall(0)
						Expect(app.GUID).To(Equal(appGUID))

						Expect(routeActor.FindOrCreateRouteCallCount()).To(BeZero())
					})
				})

				Context("when the --no-hostname flag is given", func() {
					BeforeEach(func() {
						routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("Org", "existing-app.example.com"))
						args = []string{"--no-hostname", "existing-app"}
					})

					It("binds the root domain route to an app with a pre-existing route", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(routeActor.FindOrCreateRouteCallCount()).To(Equal(1))
						hostname, _, _, _, _ := routeActor.FindOrCreateRouteArgsForCall(0)
						Expect(hostname).To(BeEmpty())
					})
				})
			})

			Context("service instances", func() {
				BeforeEach(func() {
					appRepo.CreateStub = func(params models.AppParams) (models.Application, error) {
						a := models.Application{}
						a.Name = *params.Name
						a.GUID = *params.Name + "-guid"

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

			Context("when no name and no manifest is given", func() {
				BeforeEach(func() {
					manifestRepo.ReadManifestReturns(manifest.NewEmptyManifest(), errors.New("No such manifest"))
					args = []string{}
				})

				It("fails", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr.Error()).To(ContainSubstring("Incorrect Usage. The push command requires an app name. The app name can be supplied as an argument or with a manifest.yml file."))
				})
			})
		})

		Context("when routes are specified in the manifest", func() {
			Context("and the manifest has more than one app", func() {
				BeforeEach(func() {
					m := &manifest.Manifest{
						Path: "manifest.yml",
						Data: generic.NewMap(map[interface{}]interface{}{
							"applications": []interface{}{
								generic.NewMap(map[interface{}]interface{}{
									"routes": []interface{}{
										map[interface{}]interface{}{"route": "app1route1.example.com/path"},
										map[interface{}]interface{}{"route": "app1route2.example.com:8008"},
									},
									"name": "manifest-app-name-1",
								}),
								generic.NewMap(map[interface{}]interface{}{
									"name": "manifest-app-name-2",
									"routes": []interface{}{
										map[interface{}]interface{}{"route": "app2route1.example.com"},
									},
								}),
							},
						}),
					}
					manifestRepo.ReadManifestReturns(m, nil)

					appRepo.ReadStub = func(appName string) (models.Application, error) {
						return models.Application{
							ApplicationFields: models.ApplicationFields{
								Name: appName,
								GUID: appName + "-guid",
							},
						}, nil
					}

					appRepo.UpdateStub = func(appGUID string, appParams models.AppParams) (models.Application, error) {
						return models.Application{
							ApplicationFields: models.ApplicationFields{
								GUID: appGUID,
							},
						}, nil
					}
				})

				Context("and there are no flags", func() {
					BeforeEach(func() {
						args = []string{}
					})

					It("maps the routes to the specified apps", func() {
						noHostBool := false
						emptyAppParams := models.AppParams{
							NoHostname: &noHostBool,
						}

						Expect(executeErr).ToNot(HaveOccurred())

						Expect(actor.MapManifestRouteCallCount()).To(Equal(3))

						route, app, appParams := actor.MapManifestRouteArgsForCall(0)
						Expect(route).To(Equal("app1route1.example.com/path"))
						Expect(app.ApplicationFields.GUID).To(Equal("manifest-app-name-1-guid"))
						Expect(appParams).To(Equal(emptyAppParams))

						route, app, appParams = actor.MapManifestRouteArgsForCall(1)
						Expect(route).To(Equal("app1route2.example.com:8008"))
						Expect(app.ApplicationFields.GUID).To(Equal("manifest-app-name-1-guid"))
						Expect(appParams).To(Equal(emptyAppParams))

						route, app, appParams = actor.MapManifestRouteArgsForCall(2)
						Expect(route).To(Equal("app2route1.example.com"))
						Expect(app.ApplicationFields.GUID).To(Equal("manifest-app-name-2-guid"))
						Expect(appParams).To(Equal(emptyAppParams))
					})
				})

				Context("and flags other than -f are present", func() {
					BeforeEach(func() {
						args = []string{"-n", "hostname-flag"}
					})

					It("should return an error", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr.Error()).To(Equal("Incorrect Usage. Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file."))
					})
				})
			})

			Context("and the manifest has only one app", func() {
				BeforeEach(func() {
					m := &manifest.Manifest{
						Path: "manifest.yml",
						Data: generic.NewMap(map[interface{}]interface{}{
							"applications": []interface{}{
								generic.NewMap(map[interface{}]interface{}{
									"routes": []interface{}{
										map[interface{}]interface{}{"route": "app1route1.example.com/path"},
									},
									"name": "manifest-app-name-1",
								}),
							},
						}),
					}
					manifestRepo.ReadManifestReturns(m, nil)

					appRepo.ReadStub = func(appName string) (models.Application, error) {
						return models.Application{
							ApplicationFields: models.ApplicationFields{
								Name: appName,
								GUID: appName + "-guid",
							},
						}, nil
					}

					appRepo.UpdateStub = func(appGUID string, appParams models.AppParams) (models.Application, error) {
						return models.Application{
							ApplicationFields: models.ApplicationFields{
								GUID: appGUID,
							},
						}, nil
					}
				})

				Context("and flags are present", func() {
					BeforeEach(func() {
						args = []string{"-n", "flag-value"}
					})

					It("maps the routes to the apps", func() {
						noHostBool := false
						appParamsFromContext := models.AppParams{
							Hosts:      []string{"flag-value"},
							NoHostname: &noHostBool,
						}

						Expect(executeErr).ToNot(HaveOccurred())

						Expect(actor.MapManifestRouteCallCount()).To(Equal(1))

						route, app, appParams := actor.MapManifestRouteArgsForCall(0)
						Expect(route).To(Equal("app1route1.example.com/path"))
						Expect(app.ApplicationFields.GUID).To(Equal("manifest-app-name-1-guid"))
						Expect(appParams).To(Equal(appParamsFromContext))
					})
				})
			})
		})
	})
})
