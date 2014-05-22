package service_test

import (
	"github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
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
		serviceInstance     models.ServiceInstance
		requirementsFactory *testreq.FakeReqFactory
		serviceBindingRepo  *testapi.FakeServiceBindingRepo
	)

	BeforeEach(func() {
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		serviceInstance.Name = "my-service"
		serviceInstance.Guid = "my-service-guid"

		requirementsFactory = &testreq.FakeReqFactory{}
		requirementsFactory.Application = app
		requirementsFactory.ServiceInstance = serviceInstance

		serviceBindingRepo = &testapi.FakeServiceBindingRepo{}
	})

	Context("when not logged in", func() {
		It("fails requirements when not logged in", func() {
			cmd := NewUnbindService(&testterm.FakeUI{}, testconfig.NewRepository(), serviceBindingRepo)
			testcmd.RunCommand(cmd, []string{"my-service", "my-app"}, requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when the service instance exists", func() {
			It("unbinds a service from an app", func() {
				ui := callUnbindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo)

				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding app", "my-service", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
				Expect(serviceBindingRepo.DeleteApplicationGuid).To(Equal("my-app-guid"))
			})
		})

		Context("when the service instance does not exist", func() {
			BeforeEach(func() {
				serviceBindingRepo.DeleteBindingNotFound = true
			})

			It("warns the user the the service instance does not exist", func() {
				ui := callUnbindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo)

				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding app", "my-service", "my-app"},
					[]string{"OK"},
					[]string{"my-service", "my-app", "did not exist"},
				))
				Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
				Expect(serviceBindingRepo.DeleteApplicationGuid).To(Equal("my-app-guid"))
			})
		})

		It("when no parameters are given the command fails with usage", func() {
			ui := callUnbindService([]string{"my-service"}, requirementsFactory, serviceBindingRepo)
			Expect(ui.FailedWithUsage).To(BeTrue())

			ui = callUnbindService([]string{"my-app"}, requirementsFactory, serviceBindingRepo)
			Expect(ui.FailedWithUsage).To(BeTrue())

			ui = callUnbindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo)
			Expect(ui.FailedWithUsage).To(BeFalse())
		})
	})
})

func callUnbindService(args []string, requirementsFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUnbindService(fakeUI, config, serviceBindingRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}
