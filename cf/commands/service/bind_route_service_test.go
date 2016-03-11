package service_test

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BindRouteService", func() {
	var (
		ui                      *testterm.FakeUI
		configRepo              core_config.Repository
		routeRepo               *testapi.FakeRouteRepository
		routeServiceBindingRepo *testapi.FakeRouteServiceBindingRepository

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		fakeDomain models.DomainFields

		loginRequirement           requirements.Requirement
		domainRequirement          *fakerequirements.FakeDomainRequirement
		serviceInstanceRequirement *fakerequirements.FakeServiceInstanceRequirement
		minAPIVersionRequirement   requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = &testapi.FakeRouteRepository{}
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)

		routeServiceBindingRepo = &testapi.FakeRouteServiceBindingRepository{}
		repoLocator = repoLocator.SetRouteServiceBindingRepository(routeServiceBindingRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &service.BindRouteService{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		domainRequirement = &fakerequirements.FakeDomainRequirement{}
		factory.NewDomainRequirementReturns(domainRequirement)

		fakeDomain = models.DomainFields{
			Guid: "fake-domain-guid",
			Name: "fake-domain-name",
		}
		domainRequirement.GetDomainReturns(fakeDomain)

		serviceInstanceRequirement = &fakerequirements.FakeServiceInstanceRequirement{}
		factory.NewServiceInstanceRequirementReturns(serviceInstanceRequirement)

		minAPIVersionRequirement = &passingRequirement{Name: "min-api-version-requirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires DOMAIN and SERVICE_INSTANCE as arguments"},
				))
			})
		})

		Context("when provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name", "service-instance")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a DomainRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a ServiceInstanceRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewServiceInstanceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(serviceInstanceRequirement))
			})

			It("returns a MinAPIVersionRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))

				feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(feature).To(Equal("bind-route-service"))
				expectedRequiredVersion, err := semver.Make("2.51.0")
				Expect(err).NotTo(HaveOccurred())
				Expect(requiredVersion).To(Equal(expectedRequiredVersion))
			})
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			err := flagContext.Parse("domain-name", "service-instance")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		It("tries to find the route", func() {
			cmd.Execute(flagContext)
			Expect(routeRepo.FindCallCount()).To(Equal(1))
			host, domain, path, port := routeRepo.FindArgsForCall(0)
			Expect(host).To(Equal(""))
			Expect(domain).To(Equal(fakeDomain))
			Expect(path).To(Equal(""))
			Expect(port).To(Equal(0))
		})

		Context("when given a hostname", func() {
			BeforeEach(func() {
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
				err := flagContext.Parse("domain-name", "service-instance", "-n", "the-hostname")
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to find the route with the given hostname", func() {
				cmd.Execute(flagContext)
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, _, _, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("the-hostname"))
			})
		})

		Context("when the route can be found", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{Guid: "route-guid"}, nil)
			})

			Context("when the service instance is not user-provided and requires route forwarding", func() {
				BeforeEach(func() {
					serviceInstance := models.ServiceInstance{
						ServiceOffering: models.ServiceOfferingFields{
							Requires: []string{"route_forwarding"},
						},
					}
					serviceInstance.ServicePlan = models.ServicePlanFields{
						Guid: "service-plan-guid",
					}
					serviceInstanceRequirement.GetServiceInstanceReturns(serviceInstance)
				})

				It("asks the user to confirm", func() {
					ui.Inputs = []string{"n"}
					cmd.Execute(flagContext)
					Expect(ui.Prompts).To(ContainSubstrings(
						[]string{"Binding may cause requests for route", "Do you want to proceed?"},
					))
				})

				Context("when the user confirms", func() {
					JustBeforeEach(func() {
						defer func() { recover() }()
						ui.Inputs = []string{"y"}
						cmd.Execute(flagContext)
					})

					It("does not warn", func() {
						Expect(func() []string { return ui.Outputs }).NotTo(ContainSubstrings(
							[]string{"Bind cancelled"},
						))
					})

					It("tries to bind the route service", func() {
						Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
					})

					Context("when binding the route service succeeds", func() {
						BeforeEach(func() {
							routeServiceBindingRepo.BindReturns(nil)
						})

						It("says OK", func() {
							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"OK"},
							))
						})
					})

					Context("when binding the route service fails because it is already bound", func() {
						BeforeEach(func() {
							routeServiceBindingRepo.BindReturns(errors.NewHttpError(http.StatusOK, errors.ServiceInstanceAlreadyBoundToSameRoute, "http-err"))
						})

						It("says OK", func() {
							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"OK"},
							))
						})
					})

					Context("when binding the route service fails for any other reason", func() {
						BeforeEach(func() {
							routeServiceBindingRepo.BindReturns(errors.New("bind-err"))
						})

						It("fails with the error", func() {
							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"bind-err"},
							))
						})
					})
				})

				Context("when the user does not confirm", func() {
					BeforeEach(func() {
						ui.Inputs = []string{"n"}
						cmd.Execute(flagContext)
					})

					It("warns", func() {
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Bind cancelled"},
						))
					})

					It("does not bind the route service", func() {
						Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(0))
					})
				})

				Context("when the -f flag has been passed", func() {
					BeforeEach(func() {
						flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
						flagContext.Parse("domain-name", "-f")
					})

					It("does not ask the user to confirm", func() {
						cmd.Execute(flagContext)
						Expect(ui.Prompts).NotTo(ContainSubstrings(
							[]string{"Binding may cause requests for route", "Do you want to proceed?"},
						))
					})
				})
			})

			Context("when the service instance does not require route forwarding", func() {
				BeforeEach(func() {
					serviceInstance := models.ServiceInstance{
						ServiceOffering: models.ServiceOfferingFields{
							Requires: []string{""},
						},
					}
					serviceInstanceRequirement.GetServiceInstanceReturns(serviceInstance)
				})

				It("does not ask the user to confirm", func() {
					cmd.Execute(flagContext)
					Expect(ui.Prompts).NotTo(ContainSubstrings(
						[]string{"Binding may cause requests for route", "Do you want to proceed?"},
					))
				})

				It("tells the user it is binding the route service", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Binding route", "to service instance"},
					))
				})

				It("tries to bind the route service", func() {
					cmd.Execute(flagContext)
					Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
				})

				Context("when binding the route service succeeds", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(nil)
					})

					It("says OK", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails because it is already bound", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.NewHttpError(http.StatusOK, errors.ServiceInstanceAlreadyBoundToSameRoute, "http-err"))
					})

					It("says OK", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails for any other reason", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.New("bind-err"))
					})

					It("fails with the error", func() {
						defer func() { recover() }()
						cmd.Execute(flagContext)

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"bind-err"},
						))
					})
				})
			})

			Context("when the service instance is user-provided", func() {
				BeforeEach(func() {
					serviceInstance := models.ServiceInstance{}
					serviceInstance.Guid = "service-instance-guid"
					serviceInstance.ServicePlan = models.ServicePlanFields{
						Guid: "",
					}
					serviceInstanceRequirement.GetServiceInstanceReturns(serviceInstance)
				})

				It("does not ask the user to confirm", func() {
					cmd.Execute(flagContext)
					Expect(ui.Prompts).NotTo(ContainSubstrings(
						[]string{"Binding may cause requests for route", "Do you want to proceed?"},
					))
				})

				It("tries to bind the route service", func() {
					cmd.Execute(flagContext)
					Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
					serviceInstanceGUID, routeGUID, isUserProvided, parameters := routeServiceBindingRepo.BindArgsForCall(0)
					Expect(serviceInstanceGUID).To(Equal("service-instance-guid"))
					Expect(routeGUID).To(Equal("route-guid"))
					Expect(isUserProvided).To(BeTrue())
					Expect(parameters).To(Equal(""))
				})

				Context("when given parameters as JSON", func() {
					BeforeEach(func() {
						flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
						err := flagContext.Parse("domain-name", "service-instance", "-c", `"{"some":"json"}"`)
						Expect(err).NotTo(HaveOccurred())
					})

					It("tries to find the route with the given parameters", func() {
						cmd.Execute(flagContext)
						Expect(routeRepo.FindCallCount()).To(Equal(1))
						_, _, _, parameters := routeServiceBindingRepo.BindArgsForCall(0)
						Expect(parameters).To(Equal(`{"some":"json"}`))
					})
				})

				Context("when given parameters as a file containing JSON", func() {
					BeforeEach(func() {
						flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
						tempfile, err := ioutil.TempFile("", "get-data-test")
						Expect(err).NotTo(HaveOccurred())
						jsonData := `{"some":"json"}`
						ioutil.WriteFile(tempfile.Name(), []byte(jsonData), os.ModePerm)
						err = flagContext.Parse("domain-name", "service-instance", "-c", tempfile.Name())
						Expect(err).NotTo(HaveOccurred())
					})

					It("tries to find the route with the given parameters", func() {
						cmd.Execute(flagContext)
						Expect(routeRepo.FindCallCount()).To(Equal(1))
						_, _, _, parameters := routeServiceBindingRepo.BindArgsForCall(0)
						Expect(parameters).To(Equal(`{"some":"json"}`))
					})
				})

				Context("when binding the route service succeeds", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(nil)
					})

					It("says OK", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails because it is already bound", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.NewHttpError(http.StatusOK, errors.ServiceInstanceAlreadyBoundToSameRoute, "http-err"))
					})

					It("says OK", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails for any other reason", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.New("bind-err"))
					})

					It("fails with the error", func() {
						defer func() { recover() }()
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"bind-err"},
						))
					})
				})
			})
		})

		Context("when finding the route results in an error", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{Guid: "route-guid"}, errors.New("find-err"))
			})

			It("fails with error", func() {
				defer func() { recover() }()
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"find-err"},
				))
			})
		})
	})
})
