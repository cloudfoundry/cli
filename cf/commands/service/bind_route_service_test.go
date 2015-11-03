package service_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	cferrors "github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("bind-route-service command", func() {
	var (
		ui                      *testterm.FakeUI
		requirementsFactory     *testreq.FakeReqFactory
		config                  core_config.Repository
		routeRepo               *testapi.FakeRouteRepository
		routeServiceBindingRepo *testapi.FakeRouteServiceBindingRepository
		deps                    command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetRouteServiceBindingRepository(routeServiceBindingRepo)
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("bind-route-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		routeServiceBindingRepo = &testapi.FakeRouteServiceBindingRepository{}
		routeRepo = &testapi.FakeRouteRepository{}
	})

	var callBindService = func(args []string) bool {
		return testcmd.RunCliCommand("bind-route-service", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails requirements when not logged in", func() {
		Expect(callBindService([]string{"domain", "service"})).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			routeRepo.FindReturns(models.Route{
				Guid: "some-guid",
				Host: "",
				Domain: models.DomainFields{
					Guid: "domain-guid",
					Name: "example.com",
				},
			}, nil)
		})

		It("binds a service instance to a route", func() {
			domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"

			requirementsFactory.Domain = domain
			requirementsFactory.ServiceInstance = serviceInstance
			callBindService([]string{"example.com", "my-service"})

			Expect(requirementsFactory.DomainName).To(Equal("example.com"))
			Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding route", "example.com", "my-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))
			Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
		})

		It("warns the user when the error is non HttpError ", func() {
			domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"

			requirementsFactory.Domain = domain
			requirementsFactory.ServiceInstance = serviceInstance

			routeServiceBindingRepo.BindReturns(errors.New("some-error-code"))
			callBindService([]string{"example.com", "my-service"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding route", "my-service", "example.com", "my-org", "my-space", "my-user"},
				[]string{"FAILED"},
				[]string{"some-error-code"},
			))
		})

		Context("with host name", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{
					Guid: "some-guid",
					Host: "host",
					Domain: models.DomainFields{
						Guid: "domain-guid",
						Name: "example.com",
					},
				}, nil)
			})

			It("binds a service instance to a route", func() {
				domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"

				requirementsFactory.Domain = domain
				requirementsFactory.ServiceInstance = serviceInstance
				callBindService([]string{"example.com", "my-service", "-n", "host"})

				Expect(requirementsFactory.DomainName).To(Equal("example.com"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Binding route", "host.example.com", "my-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
			})
		})

		Context("when route repo returns error", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{}, errors.New("kaboom"))
			})

			It("displays the error returned by route repo", func() {
				domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"

				requirementsFactory.Domain = domain
				requirementsFactory.ServiceInstance = serviceInstance
				callBindService([]string{"example.com", "my-service", "-n", "host"})

				Expect(requirementsFactory.DomainName).To(Equal("example.com"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"kaboom"},
				))
				Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(0))
			})
		})

		Context("when route service bind repo returns an error", func() {
			BeforeEach(func() {
				routeServiceBindingRepo.BindReturns(errors.New("bind-error"))
			})

			It("displays the error returned by binding repo", func() {
				domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"

				requirementsFactory.Domain = domain
				requirementsFactory.ServiceInstance = serviceInstance
				callBindService([]string{"example.com", "my-service", "-n", "host"})

				Expect(requirementsFactory.DomainName).To(Equal("example.com"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"bind-error"},
				))
			})

			Context("when route service bind repo returns 130008", func() {
				BeforeEach(func() {
					httpErr := cferrors.NewHttpError(200, "130008", "baam!")
					routeServiceBindingRepo.BindReturns(httpErr)
				})

				It("returns OK and displays a warning", func() {
					domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

					serviceInstance := models.ServiceInstance{}
					serviceInstance.Name = "my-service"
					serviceInstance.Guid = "my-service-guid"

					requirementsFactory.Domain = domain
					requirementsFactory.ServiceInstance = serviceInstance
					callBindService([]string{"example.com", "my-service"})

					Expect(requirementsFactory.DomainName).To(Equal("example.com"))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Route", "is already bound to service", "example.com", "my-service"},
						[]string{"OK"},
					))
				})
			})
		})

		It("fails with usage when called without a service instance and route", func() {
			callBindService([]string{"my-service"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))

			ui = &testterm.FakeUI{}
			callBindService([]string{"example.com"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))

			ui = &testterm.FakeUI{}
			callBindService([]string{"my-service", "example.com"})
			Expect(ui.Outputs).ToNot(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		Context("when service instance requires route_forwarding", func() {
			BeforeEach(func() {
				domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"
				serviceInstance.ServicePlan.Guid = "managed"
				serviceInstance.ServiceOffering.Requires = []string{"route_forwarding"}

				requirementsFactory.Domain = domain
				requirementsFactory.ServiceInstance = serviceInstance
			})

			It("prompts for confirmation", func() {
				ui.Inputs = []string{"yes"}
				callBindService([]string{"example.com", "my-service"})
				Expect(len(ui.Prompts)).To(Equal(1))
				Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
				Expect(ui.Prompts).To(ContainSubstrings(
					[]string{"Binding may cause requests for route", "example.com", "to be altered by service", "my-service"},
				))
			})

			Context("when the user says no", func() {
				BeforeEach(func() {
					ui.Inputs = []string{"no"}
				})

				It("does not call bind", func() {
					callBindService([]string{"example.com", "my-service"})
					Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(0))
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Bind cancelled"},
					))
				})
			})

			Context("when -f flag is provided", func() {
				It("binds a service instance to a route without confirmation", func() {
					callBindService([]string{"example.com", "my-service", "-f"})
					Expect(len(ui.Prompts)).To(Equal(0))
					Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
				})
			})
		})

		Context("when service instance is not a managed service (UPSI)", func() {
			BeforeEach(func() {
				domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"

				requirementsFactory.Domain = domain
				requirementsFactory.ServiceInstance = serviceInstance
			})

			It("binds a service instance to a route without confirmation", func() {
				callBindService([]string{"example.com", "my-service"})
				Expect(len(ui.Prompts)).To(Equal(0))
				Expect(routeServiceBindingRepo.BindCallCount()).To(Equal(1))
			})
		})
	})
})
