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

var _ = Describe("unbind-route-service command", func() {
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
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("unbind-route-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		routeServiceBindingRepo = &testapi.FakeRouteServiceBindingRepository{}
		routeRepo = &testapi.FakeRouteRepository{}
	})

	var callUnbindService = func(args []string) bool {
		return testcmd.RunCliCommand("unbind-route-service", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails requirements when not logged in", func() {
		Expect(callUnbindService([]string{"domain", "service"})).To(BeFalse())
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
			domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"

			requirementsFactory.Domain = domain
			requirementsFactory.ServiceInstance = serviceInstance
		})

		It("fails with usage when called without a service instance and route", func() {
			callUnbindService([]string{"my-service"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))

			ui = &testterm.FakeUI{}
			callUnbindService([]string{"example.com"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		Context("when user says yes", func() {
			It("prompts for confirmation and unbinds the service", func() {
				ui.Inputs = []string{"yes"}
				callUnbindService([]string{"example.com", "my-service"})
				Expect(len(ui.Prompts)).To(Equal(1))
				Expect(routeServiceBindingRepo.UnbindCallCount()).To(Equal(1))
				Expect(ui.Prompts).To(ContainSubstrings(
					[]string{"Unbinding may leave apps mapped to route", "example.com", "vulnerable; e.g. if service instance", "my-service", "provides authentication. Do you want to proceed?"},
				))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding route", "example.com", "from service instance", "my-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
			})
		})

		Context("when the user says no", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"no"}
			})

			It("does not call unbind", func() {
				callUnbindService([]string{"example.com", "my-service"})
				Expect(len(ui.Prompts)).To(Equal(1))
				Expect(routeServiceBindingRepo.UnbindCallCount()).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbind cancelled"},
				))
			})
		})

		Context("when -f flag is provided", func() {
			It("unbinds a service instance to a route without confirmation", func() {
				callUnbindService([]string{"example.com", "my-service", "-f"})
				Expect(len(ui.Prompts)).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding route", "example.com", "from service instance", "my-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(routeServiceBindingRepo.UnbindCallCount()).To(Equal(1))
			})
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
				ui.Inputs = []string{"yes"}
			})

			It("binds a service instance to a route", func() {
				domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"

				requirementsFactory.Domain = domain
				requirementsFactory.ServiceInstance = serviceInstance
				callUnbindService([]string{"example.com", "my-service", "-n", "host"})

				Expect(requirementsFactory.DomainName).To(Equal("example.com"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(len(ui.Prompts)).To(Equal(1))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding route", "host.example.com", "my-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(routeServiceBindingRepo.UnbindCallCount()).To(Equal(1))
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
				callUnbindService([]string{"example.com", "my-service", "-n", "host", "-f"})

				Expect(requirementsFactory.DomainName).To(Equal("example.com"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(len(ui.Prompts)).To(Equal(0))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"kaboom"},
				))
				Expect(routeServiceBindingRepo.UnbindCallCount()).To(Equal(0))
			})
		})

		Context("when route service bind repo returns an error", func() {
			BeforeEach(func() {
				routeServiceBindingRepo.UnbindReturns(errors.New("unbind-error"))
			})

			It("displays the error returned by binding repo", func() {
				domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"

				requirementsFactory.Domain = domain
				requirementsFactory.ServiceInstance = serviceInstance
				callUnbindService([]string{"example.com", "my-service", "-n", "host", "-f"})
				Expect(len(ui.Prompts)).To(Equal(0))

				Expect(requirementsFactory.DomainName).To(Equal("example.com"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"unbind-error"},
				))
			})
		})

		Context("when route service bind repo returns an error", func() {
			BeforeEach(func() {
				routeServiceBindingRepo.UnbindReturns(cferrors.NewHttpError(400, "1002", "Sorry We're Not Related"))
			})

			It("displays a warning and return okay", func() {
				domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"

				requirementsFactory.Domain = domain
				requirementsFactory.ServiceInstance = serviceInstance
				callUnbindService([]string{"example.com", "my-service", "-n", "host", "-f"})
				Expect(len(ui.Prompts)).To(Equal(0))

				Expect(requirementsFactory.DomainName).To(Equal("example.com"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"OK"},
					[]string{"Route example.com was not bound to service instance my-service"},
				))
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

			It("unbinds a service instance to a route with confirmation", func() {
				callUnbindService([]string{"example.com", "my-service", "-f"})
				Expect(len(ui.Prompts)).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding route", "example.com", "my-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(routeServiceBindingRepo.UnbindCallCount()).To(Equal(1))
			})
		})
	})
})
