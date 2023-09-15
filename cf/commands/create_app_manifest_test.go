package commands_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/manifest/manifestfakes"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/stacks/stacksfakes"
	testconfig "code.cloudfoundry.org/cli/cf/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/cf/util/testhelpers/terminal"

	"os"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateAppManifest", func() {
	var (
		appName        string
		ui             *testterm.FakeUI
		configRepo     coreconfig.Repository
		appSummaryRepo *apifakes.FakeAppSummaryRepository
		stackRepo      *stacksfakes.FakeStackRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
		applicationRequirement   *requirementsfakes.FakeApplicationRequirement

		fakeManifest *manifestfakes.FakeApp
	)

	BeforeEach(func() {
		rand, err := uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())
		appName = fmt.Sprintf("app-name-%s", rand)
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		appSummaryRepo = new(apifakes.FakeAppSummaryRepository)
		repoLocator := deps.RepoLocator.SetAppSummaryRepository(appSummaryRepo)
		stackRepo = new(stacksfakes.FakeStackRepository)
		repoLocator = repoLocator.SetStackRepository(stackRepo)

		fakeManifest = new(manifestfakes.FakeApp)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
			AppManifest: fakeManifest,
		}

		cmd = &commands.CreateAppManifest{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedSpaceRequirement = &passingRequirement{Name: "targeted-space-requirement"}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)

		applicationRequirement = new(requirementsfakes.FakeApplicationRequirement)
		application := models.Application{}
		application.GUID = "app-guid"
		applicationRequirement.GetApplicationReturns(application)
		factory.NewApplicationRequirementReturns(applicationRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse(appName, "extra-arg")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires APP_NAME as argument"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse(appName)
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns an ApplicationRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualRequirements).To(ContainElement(applicationRequirement))
			})

			It("returns a TargetedSpaceRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var (
			application models.Application
			runCLIErr   error
		)

		BeforeEach(func() {
			err := flagContext.Parse(appName)
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)

			application = models.Application{}
			application.Name = appName
		})

		JustBeforeEach(func() {
			runCLIErr = cmd.Execute(flagContext)
		})

		AfterEach(func() {
			os.Remove(fmt.Sprintf("%s_manifest.yml", appName))
		})

		Context("when there is an app summary", func() {
			BeforeEach(func() {
				appSummaryRepo.GetSummaryReturns(application, nil)
			})

			It("tries to get the app summary", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(appSummaryRepo.GetSummaryCallCount()).To(Equal(1))
			})
		})

		Context("when there is an error getting the app summary", func() {
			BeforeEach(func() {
				appSummaryRepo.GetSummaryReturns(models.Application{}, errors.New("get-summary-err"))
			})

			It("prints an error", func() {
				Expect(runCLIErr).To(HaveOccurred())
				Expect(runCLIErr.Error()).To(Equal("Error getting application summary: get-summary-err"))
			})
		})

		Context("when getting the app summary succeeds", func() {
			BeforeEach(func() {
				application.Memory = 1024
				application.InstanceCount = 2
				application.StackGUID = "the-stack-guid"
				appSummaryRepo.GetSummaryReturns(application, nil)
			})

			It("sets memory", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(fakeManifest.MemoryCallCount()).To(Equal(1))
				name, memory := fakeManifest.MemoryArgsForCall(0)
				Expect(name).To(Equal(appName))
				Expect(memory).To(Equal(int64(1024)))
			})

			It("sets instances", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(fakeManifest.InstancesCallCount()).To(Equal(1))
				name, instances := fakeManifest.InstancesArgsForCall(0)
				Expect(name).To(Equal(appName))
				Expect(instances).To(Equal(2))
			})

			It("tries to get stacks", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(stackRepo.FindByGUIDCallCount()).To(Equal(1))
				Expect(stackRepo.FindByGUIDArgsForCall(0)).To(Equal("the-stack-guid"))
			})

			Context("when getting stacks succeeds", func() {
				BeforeEach(func() {
					stackRepo.FindByGUIDReturns(models.Stack{
						GUID: "the-stack-guid",
						Name: "the-stack-name",
					}, nil)
				})

				It("sets the stacks", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(fakeManifest.StackCallCount()).To(Equal(1))
					name, stackName := fakeManifest.StackArgsForCall(0)
					Expect(name).To(Equal(appName))
					Expect(stackName).To(Equal("the-stack-name"))
				})
			})

			Context("when getting stacks fails", func() {
				BeforeEach(func() {
					stackRepo.FindByGUIDReturns(models.Stack{}, errors.New("find-by-guid-err"))
				})

				It("fails with error", func() {
					Expect(runCLIErr).To(HaveOccurred())
					Expect(runCLIErr.Error()).To(Equal("Error retrieving stack: find-by-guid-err"))
				})
			})

			It("tries to save the manifest", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(fakeManifest.SaveCallCount()).To(Equal(1))
			})

			Context("when saving the manifest succeeds", func() {
				BeforeEach(func() {
					fakeManifest.SaveReturns(nil)
				})

				It("says OK", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"OK"},
						[]string{fmt.Sprintf("Manifest file created successfully at ./%s_manifest.yml", appName)},
					))
				})
			})

			Context("when saving the manifest fails", func() {
				BeforeEach(func() {
					fakeManifest.SaveReturns(errors.New("save-err"))
				})

				It("fails with error", func() {
					Expect(runCLIErr).To(HaveOccurred())
					Expect(runCLIErr.Error()).To(Equal("Error creating manifest file: save-err"))
				})
			})

			Context("when the app has a command", func() {
				BeforeEach(func() {
					application.Command = "app-command"
					appSummaryRepo.GetSummaryReturns(application, nil)
				})

				It("sets the start command", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(fakeManifest.StartCommandCallCount()).To(Equal(1))
					name, command := fakeManifest.StartCommandArgsForCall(0)
					Expect(name).To(Equal(appName))
					Expect(command).To(Equal("app-command"))
				})
			})

			Context("when the app has a buildpack", func() {
				BeforeEach(func() {
					application.BuildpackURL = "buildpack"
					appSummaryRepo.GetSummaryReturns(application, nil)
				})

				It("sets the buildpack", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(fakeManifest.BuildpackURLCallCount()).To(Equal(1))
					name, buildpack := fakeManifest.BuildpackURLArgsForCall(0)
					Expect(name).To(Equal(appName))
					Expect(buildpack).To(Equal("buildpack"))
				})
			})

			Context("when the app has services", func() {
				BeforeEach(func() {
					application.Services = []models.ServicePlanSummary{
						{
							Name: "sp1-name",
						},
						{
							Name: "sp2-name",
						},
					}
					appSummaryRepo.GetSummaryReturns(application, nil)
				})

				It("sets the services", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(fakeManifest.ServiceCallCount()).To(Equal(2))

					name, service := fakeManifest.ServiceArgsForCall(0)
					Expect(name).To(Equal(appName))
					Expect(service).To(Equal("sp1-name"))

					name, service = fakeManifest.ServiceArgsForCall(1)
					Expect(name).To(Equal(appName))
					Expect(service).To(Equal("sp2-name"))
				})
			})

			Context("when the app has a health check timeout", func() {
				BeforeEach(func() {
					application.HealthCheckTimeout = 5
					appSummaryRepo.GetSummaryReturns(application, nil)
				})

				It("sets the health check timeout", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(fakeManifest.HealthCheckTimeoutCallCount()).To(Equal(1))
					name, timeout := fakeManifest.HealthCheckTimeoutArgsForCall(0)
					Expect(name).To(Equal(appName))
					Expect(timeout).To(Equal(5))
				})
			})

			Context("when the app has environment vars", func() {
				BeforeEach(func() {
					application.EnvironmentVars = map[string]interface{}{
						"float64-key": float64(5),
						"bool-key":    true,
						"string-key":  "string",
					}
					appSummaryRepo.GetSummaryReturns(application, nil)
				})

				It("sets the env vars", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(fakeManifest.EnvironmentVarsCallCount()).To(Equal(3))
					actuals := map[string]interface{}{}

					for i := 0; i < 3; i++ {
						name, k, v := fakeManifest.EnvironmentVarsArgsForCall(i)
						Expect(name).To(Equal(appName))
						actuals[k] = v
					}

					Expect(actuals["float64-key"]).To(Equal("5"))
					Expect(actuals["bool-key"]).To(Equal("true"))
					Expect(actuals["string-key"]).To(Equal("string"))
				})
			})

			Context("when the app has an environment var of an unsupported type", func() {
				BeforeEach(func() {
					application.EnvironmentVars = map[string]interface{}{
						"key": int(1),
					}
					appSummaryRepo.GetSummaryReturns(application, nil)
				})

				It("fails with error", func() {
					Expect(runCLIErr).To(HaveOccurred())
					Expect(runCLIErr.Error()).To(Equal("Failed to create manifest, unable to parse environment variable: key"))
				})
			})

			Context("when the app has routes", func() {
				BeforeEach(func() {
					application.Routes = []models.RouteSummary{
						{
							Host: "route-1-host",
							Domain: models.DomainFields{
								Name: "http-domain",
							},
							Path: "path",
							Port: 0,
						},
						{
							Host: "",
							Domain: models.DomainFields{
								Name: "tcp-domain",
							},
							Path: "",
							Port: 123,
						},
					}
					appSummaryRepo.GetSummaryReturns(application, nil)
				})

				It("sets the domains", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(fakeManifest.RouteCallCount()).To(Equal(2))

					name, host, domainName, path, port := fakeManifest.RouteArgsForCall(0)
					Expect(name).To(Equal(appName))
					Expect(host).To(Equal("route-1-host"))
					Expect(domainName).To(Equal("http-domain"))
					Expect(path).To(Equal("path"))
					Expect(port).To(Equal(0))

					name, host, domainName, path, port = fakeManifest.RouteArgsForCall(1)
					Expect(name).To(Equal(appName))
					Expect(host).To(Equal(""))
					Expect(domainName).To(Equal("tcp-domain"))
					Expect(path).To(Equal(""))
					Expect(port).To(Equal(123))
				})
			})

			Context("when the app has a disk quota", func() {
				BeforeEach(func() {
					application.DiskQuota = 1024
					appSummaryRepo.GetSummaryReturns(application, nil)
				})

				It("sets the disk quota", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(fakeManifest.DiskQuotaCallCount()).To(Equal(1))
					name, quota := fakeManifest.DiskQuotaArgsForCall(0)
					Expect(name).To(Equal(appName))
					Expect(quota).To(Equal(int64(1024)))
				})
			})

			Context("when the app has health check type", func() {
				Context("when the health check type is port", func() {
					BeforeEach(func() {
						application.HealthCheckType = "port"
						appSummaryRepo.GetSummaryReturns(application, nil)
					})

					It("does not set the health check type nor the endpoint", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())

						Expect(fakeManifest.HealthCheckTypeCallCount()).To(Equal(0))
						Expect(fakeManifest.HealthCheckHTTPEndpointCallCount()).To(Equal(0))
					})
				})

				Context("when the health check type is process", func() {
					BeforeEach(func() {
						application.HealthCheckType = "process"
						appSummaryRepo.GetSummaryReturns(application, nil)
					})

					It("sets the health check type but not the endpoint", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())

						Expect(fakeManifest.HealthCheckTypeCallCount()).To(Equal(1))
						name, healthCheckType := fakeManifest.HealthCheckTypeArgsForCall(0)
						Expect(name).To(Equal(appName))
						Expect(healthCheckType).To(Equal("process"))
						Expect(fakeManifest.HealthCheckHTTPEndpointCallCount()).To(Equal(0))
					})
				})

				Context("when the health check type is none", func() {
					BeforeEach(func() {
						application.HealthCheckType = "none"
						appSummaryRepo.GetSummaryReturns(application, nil)
					})

					It("sets the health check type but not the endpoint", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())

						Expect(fakeManifest.HealthCheckTypeCallCount()).To(Equal(1))
						name, healthCheckType := fakeManifest.HealthCheckTypeArgsForCall(0)
						Expect(name).To(Equal(appName))
						Expect(healthCheckType).To(Equal("none"))
						Expect(fakeManifest.HealthCheckHTTPEndpointCallCount()).To(Equal(0))
					})
				})

				Context("when the health check type is http", func() {
					BeforeEach(func() {
						application.HealthCheckType = "http"
						appSummaryRepo.GetSummaryReturns(application, nil)
					})

					It("sets the health check type", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())

						Expect(fakeManifest.HealthCheckTypeCallCount()).To(Equal(1))
						name, healthCheckType := fakeManifest.HealthCheckTypeArgsForCall(0)
						Expect(name).To(Equal(appName))
						Expect(healthCheckType).To(Equal("http"))
					})

					Context("when the health check endpoint is the empty string", func() {
						BeforeEach(func() {
							application.HealthCheckHTTPEndpoint = ""
							appSummaryRepo.GetSummaryReturns(application, nil)
						})

						It("does not set the health check endpoint", func() {
							Expect(runCLIErr).NotTo(HaveOccurred())

							Expect(fakeManifest.HealthCheckHTTPEndpointCallCount()).To(Equal(0))
						})
					})

					Context("when the health check endpoint is /", func() {
						BeforeEach(func() {
							application.HealthCheckHTTPEndpoint = "/"
							appSummaryRepo.GetSummaryReturns(application, nil)
						})

						It("does not set the health check endpoint", func() {
							Expect(runCLIErr).NotTo(HaveOccurred())

							Expect(fakeManifest.HealthCheckHTTPEndpointCallCount()).To(Equal(0))
						})
					})

					Context("when the health check endpoint is not /", func() {
						BeforeEach(func() {
							application.HealthCheckHTTPEndpoint = "/some-endpoint"
							appSummaryRepo.GetSummaryReturns(application, nil)
						})

						It("sets the health check endpoint", func() {
							Expect(runCLIErr).NotTo(HaveOccurred())

							Expect(fakeManifest.HealthCheckHTTPEndpointCallCount()).To(Equal(1))
							name, healthCheckHTTPEndpoint := fakeManifest.HealthCheckHTTPEndpointArgsForCall(0)
							Expect(name).To(Equal(appName))
							Expect(healthCheckHTTPEndpoint).To(Equal("/some-endpoint"))
						})
					})
				})
			})
		})
	})
})
