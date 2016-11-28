package service_test

import (
	"net/http"

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

var _ = Describe("UnbindRouteService", func() {
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
		minAPIVersionRequirement   requirements.Requirement
		domainRequirement          *requirementsfakes.FakeDomainRequirement
		serviceInstanceRequirement *requirementsfakes.FakeServiceInstanceRequirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

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

		cmd = &service.UnbindRouteService{}
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
				Expect(feature).To(Equal("unbind-route-service"))
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
			ui.Inputs = []string{"n"}
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
				ui.Inputs = []string{"n"}
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
				ui.Inputs = []string{"n"}
			})

			It("should attempt to find the route", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				_, _, path, _ := routeRepo.FindArgsForCall(0)
				Expect(path).To(Equal("/path"))
			})

			Context("when the path does not contain a leading slash", func() {
				BeforeEach(func() {
					flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
					err := flagContext.Parse("domain-name", "service-instance", "--path", "path")
					Expect(err).NotTo(HaveOccurred())
					ui.Inputs = []string{"n"}
				})

				It("should prefix the path with a leading slash and attempt to find the route", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(routeRepo.FindCallCount()).To(Equal(1))
					_, _, path, _ := routeRepo.FindArgsForCall(0)
					Expect(path).To(Equal("/path"))
				})
			})
		})

		Context("when given hostname and path", func() {
			BeforeEach(func() {
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
				err := flagContext.Parse("domain-name", "service-instance", "--hostname", "the-hostname", "--path", "path")
				Expect(err).NotTo(HaveOccurred())
				ui.Inputs = []string{"n"}
			})

			It("should attempt to find the route", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				hostname, _, path, _ := routeRepo.FindArgsForCall(0)
				Expect(hostname).To(Equal("the-hostname"))
				Expect(path).To(Equal("/path"))
			})
		})

		Context("when the route can be found", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{GUID: "route-guid"}, nil)
				ui.Inputs = []string{"n"}
			})

			It("asks the user to confirm", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(ui.Prompts).To(ContainSubstrings(
					[]string{"Unbinding may leave apps mapped to route", "Do you want to proceed?"},
				))
			})

			Context("when the user confirms", func() {
				BeforeEach(func() {
					ui.Inputs = []string{"y"}
				})

				It("does not warn", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(func() []string {
						return ui.Outputs()
					}).NotTo(ContainSubstrings(
						[]string{"Unbind cancelled"},
					))
				})

				It("tells the user it is unbinding the route service", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Unbinding route", "from service instance"},
					))
				})

				It("tries to unbind the route service", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(routeServiceBindingRepo.UnbindCallCount()).To(Equal(1))
				})

				Context("when unbinding the route service succeeds", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.UnbindReturns(nil)
					})

					It("says OK", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when unbinding the route service fails because it was not bound", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.UnbindReturns(errors.NewHTTPError(http.StatusOK, errors.InvalidRelation, "http-err"))
					})

					It("says OK", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})

					It("warns", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Route", "was not bound to service instance"},
						))
					})
				})

				Context("when unbinding the route service fails for any other reason", func() {
					BeforeEach(func() {
						routeServiceBindingRepo.UnbindReturns(errors.New("unbind-err"))
					})

					It("fails with the error", func() {
						Expect(runCLIErr).To(HaveOccurred())
						Expect(runCLIErr.Error()).To(Equal("unbind-err"))
					})
				})
			})

			Context("when the user does not confirm", func() {
				BeforeEach(func() {
					ui.Inputs = []string{"n"}
				})

				It("warns", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Unbind cancelled"},
					))
				})

				It("does not bind the route service", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(routeServiceBindingRepo.UnbindCallCount()).To(Equal(0))
				})
			})

			Context("when the -f flag has been passed", func() {
				BeforeEach(func() {
					flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
					flagContext.Parse("domain-name", "-f")
				})

				It("does not ask the user to confirm", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(ui.Prompts).NotTo(ContainSubstrings(
						[]string{"Unbinding may leave apps mapped to route", "Do you want to proceed?"},
					))
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
