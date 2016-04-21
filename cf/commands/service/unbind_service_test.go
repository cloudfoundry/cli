package service_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("unbind-service command", func() {
	var (
		app                 models.Application
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		serviceInstance     models.ServiceInstance
		requirementsFactory *testreq.FakeReqFactory
		serviceBindingRepo  *apifakes.OldFakeServiceBindingRepo
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceBindingRepository(serviceBindingRepo)
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("unbind-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		app.Name = "my-app"
		app.GUID = "my-app-guid"

		ui = &testterm.FakeUI{}
		serviceInstance.Name = "my-service"
		serviceInstance.GUID = "my-service-guid"

		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		requirementsFactory.Application = app
		requirementsFactory.ServiceInstance = serviceInstance

		serviceBindingRepo = new(apifakes.OldFakeServiceBindingRepo)
	})

	callUnbindService := func(args []string) bool {
		return testcmd.RunCLICommand("unbind-service", args, requirementsFactory, updateCommandDependency, false)
	}

	Context("when not logged in", func() {
		It("fails requirements when not logged in", func() {
			Expect(testcmd.RunCLICommand("unbind-service", []string{"my-service", "my-app"}, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when the service instance exists", func() {
			It("unbinds a service from an app", func() {
				callUnbindService([]string{"my-app", "my-service"})

				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding app", "my-service", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
				Expect(serviceBindingRepo.DeleteApplicationGUID).To(Equal("my-app-guid"))
			})
		})

		Context("when the service instance does not exist", func() {
			BeforeEach(func() {
				serviceBindingRepo.DeleteBindingNotFound = true
			})

			It("warns the user the the service instance does not exist", func() {
				callUnbindService([]string{"my-app", "my-service"})

				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding app", "my-service", "my-app"},
					[]string{"OK"},
					[]string{"my-service", "my-app", "did not exist"},
				))
				Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
				Expect(serviceBindingRepo.DeleteApplicationGUID).To(Equal("my-app-guid"))
			})
		})

		It("when no parameters are given the command fails with usage", func() {
			callUnbindService([]string{"my-service"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))

			ui = &testterm.FakeUI{}
			callUnbindService([]string{"my-app"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))

			ui = &testterm.FakeUI{}
			callUnbindService([]string{"my-app", "my-service"})
			Expect(ui.Outputs).ToNot(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})
	})
})
