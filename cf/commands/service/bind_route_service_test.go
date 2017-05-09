package service_test

import (
	"io/ioutil"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"github.com/blang/semver"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BindRouteService", func() {
	var (
		ui                      *testterm.FakeUI
		configRepo              coreconfig.Repository
		routeRepo               *apifakes.FakeRouteRepository
		routeServiceBindingRepo *apifakes.FakeRouteServiceBindingRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		fakeDomain models.DomainFields

		loginRequirement           requirements.Requirement
		domainRequirement          *requirementsfakes.FakeDomainRequirement
		serviceInstanceRequirement *requirementsfakes.FakeServiceInstanceRequirement
		minAPIVersionRequirement   requirements.Requirement
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)

		configRepo = testconfig.NewRepositoryWithDefaults()

		routeRepo = new(apifakes.FakeRouteRepository)
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)

		routeServiceBindingRepo = new(apifakes.FakeRouteServiceBindingRepository)
		repoLocator = repoLocator.SetRouteServiceBindingRepository(routeServiceBindingRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = new(service.BindRouteService)
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		domainRequirement = new(requirementsfakes.FakeDomainRequirement)
		factory.NewDomainRequirementReturns(domainRequirement)

		fakeDomain = models.DomainFields{
			GUID: "fake-domain-guid",
			Name: "fake-domain-name",
		}
		domainRequirement.GetDomainReturns(fakeDomain)

		serviceInstanceRequirement = new(requirementsfakes.FakeServiceInstanceRequirement)
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
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a DomainRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a ServiceInstanceRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewServiceInstanceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(serviceInstanceRequirement))
			})

			It("returns a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
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
		var runCLIErr error

		BeforeEach(func() {
			err := flagContext.Parse("domain-name", "service-instance")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		JustBeforeEach(func() {
			runCLIErr = cmd.Execute(flagContext)
		})

		It("tries to find the route", func() {
			Expect(runCLIErr).NotTo(HaveOccurred())
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
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, _, _, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("the-hostname"))
			})
		})

		Context("when given a path", func() {
			BeforeEach(func() {
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
				err := flagContext.Parse("domain-name", "service-instance", "--path", "/path")
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to find the route with the given path", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				_, _, path, _ := routeRepo.FindArgsForCall(0)
				Expect(path).To(Equal("/path"))
			})

			Context("when the path does not have a leading slash", func() {
				BeforeEach(func() {
					flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
					err := flagContext.Parse("domain-name", "service-instance", "--path", "path")
					Expect(err).NotTo(HaveOccurred())
				})

				It("prepends a leading slash and tries to find the route with the given path", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(routeRepo.FindCallCount()).To(Equal(1))
					_, _, path, _ := routeRepo.FindArgsForCall(0)
					Expect(path).To(Equal("/path"))
				})
			})
		})

		Context("when given a path and a hostname", func() {
			BeforeEach(func() {
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
				err := flagContext.Parse("domain-name", "service-instance", "-n", "the-hostname", "--path", "/path")
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to find the route with both the given hostname and path", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				host, _, path, _ := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("the-hostname"))
				Expect(path).To(Equal("/path"))
			})
		})

		Context("when the route can be found", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{GUID: "route-guid"}, nil)
			})

			Context("when the service instance is not user-provided and requires route forwarding", func() {
				BeforeEach(func() {
					serviceInstance := models.ServiceInstance{
						ServiceOffering: models.ServiceOfferingFields{
							Requires: []string{"route_forwarding"},
						},
					}
					serviceInstance.ServicePlan = models.ServicePlanFields{
						GUID: "service-plan-guid",
					}
					serviceInstanceRequirement.GetServiceInstanceReturns(serviceInstance)
				})

				It("does not warn", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).NotTo(ContainSubstrings(
						[]string{"Bind cancelled"},
					))
				})

				It("tries to bind the route service", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
				})

				Context("when binding the route service succeeds", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(nil)
					})

					It("says OK", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails because it is already bound", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.NewHTTPError(http.StatusOK, errors.ServiceInstanceAlreadyBoundToSameRoute, "http-err"))
					})

					It("says OK", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails for any other reason", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.New("bind-err"))
					})

					It("fails with the error", func() {
						Expect(runCLIErr).To(HaveOccurred())
						Expect(runCLIErr.Error()).To(Equal("bind-err"))
					})
				})

				Context("when the -f flag has been passed", func() {
					BeforeEach(func() {
						flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
					})

					It("does not alter the behavior", func() {
						err := flagContext.Parse("domain-name", "-f")
						Expect(err).NotTo(HaveOccurred())

						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
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
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Prompts).NotTo(ContainSubstrings(
						[]string{"Binding may cause requests for route", "Do you want to proceed?"},
					))
				})

				It("tells the user it is binding the route service", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Binding route", "to service instance"},
					))
				})

				It("tries to bind the route service", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
				})

				Context("when binding the route service succeeds", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(nil)
					})

					It("says OK", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails because it is already bound", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.NewHTTPError(http.StatusOK, errors.ServiceInstanceAlreadyBoundToSameRoute, "http-err"))
					})

					It("says OK", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails for any other reason", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.New("bind-err"))
					})

					It("fails with the error", func() {
						Expect(runCLIErr).To(HaveOccurred())
						Expect(runCLIErr.Error()).To(Equal("bind-err"))
					})
				})
			})

			Context("when the service instance is user-provided", func() {
				BeforeEach(func() {
					serviceInstance := models.ServiceInstance{}
					serviceInstance.GUID = "service-instance-guid"
					serviceInstance.ServicePlan = models.ServicePlanFields{
						GUID: "",
					}
					serviceInstanceRequirement.GetServiceInstanceReturns(serviceInstance)
				})

				It("does not ask the user to confirm", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Prompts).NotTo(ContainSubstrings(
						[]string{"Binding may cause requests for route", "Do you want to proceed?"},
					))
				})

				It("tries to bind the route service", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
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
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(routeRepo.FindCallCount()).To(Equal(1))
						_, _, _, parameters := routeServiceBindingRepo.BindArgsForCall(0)
						Expect(parameters).To(Equal(`{"some":"json"}`))
					})
				})

				Context("when given parameters as a file containing JSON", func() {
					var filename string
					BeforeEach(func() {
						flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
						tempfile, err := ioutil.TempFile("", "get-data-test")
						Expect(err).NotTo(HaveOccurred())
						Expect(tempfile.Close()).NotTo(HaveOccurred())
						filename = tempfile.Name()

						jsonData := `{"some":"json"}`
						ioutil.WriteFile(filename, []byte(jsonData), os.ModePerm)
						err = flagContext.Parse("domain-name", "service-instance", "-c", filename)
						Expect(err).NotTo(HaveOccurred())
					})

					AfterEach(func() {
						os.RemoveAll(filename)
					})

					It("tries to find the route with the given parameters", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
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
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails because it is already bound", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.NewHTTPError(http.StatusOK, errors.ServiceInstanceAlreadyBoundToSameRoute, "http-err"))
					})

					It("says OK", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when binding the route service fails for any other reason", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.BindReturns(errors.New("bind-err"))
					})

					It("fails with the error", func() {
						Expect(runCLIErr).To(HaveOccurred())
						Expect(runCLIErr.Error()).To(Equal("bind-err"))
					})
				})
			})
		})

		Context("when finding the route results in an error", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{GUID: "route-guid"}, errors.New("find-err"))
			})

			It("fails with error", func() {
				Expect(runCLIErr).To(HaveOccurred())
				Expect(runCLIErr.Error()).To(Equal("find-err"))
			})
		})
	})
})
