package service_test

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/applications"
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
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
		serviceInstance     models.ServiceInstance
		requirementsFactory *testreq.FakeReqFactory
		serviceBindingRepo  *testapi.FakeServiceBindingRepo
		stagingWatcher      *fakeStagingWatcher
		appRepo             *testApplication.FakeApplicationRepository
		configRepo          core_config.ReadWriter
	)

	BeforeEach(func() {
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		serviceInstance.Name = "my-service"
		serviceInstance.Guid = "my-service-guid"

		requirementsFactory = &testreq.FakeReqFactory{}
		requirementsFactory.Application = app
		requirementsFactory.ServiceInstance = serviceInstance

		appRepo = &testApplication.FakeApplicationRepository{}
		stagingWatcher = &fakeStagingWatcher{}

		serviceBindingRepo = &testapi.FakeServiceBindingRepo{}
	})

	Context("when not logged in", func() {
		It("fails requirements when not logged in", func() {
			cmd := NewUnbindService(&testterm.FakeUI{}, testconfig.NewRepository(), serviceBindingRepo, appRepo, stagingWatcher)
			Expect(testcmd.RunCommand(cmd, []string{"my-service", "my-app"}, requirementsFactory)).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			configRepo = testconfig.NewRepositoryWithDefaults()
		})

		Context("when the service instance exists", func() {
			It("unbinds a service from an app", func() {
				ui := callUnbindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)

				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding app", "my-service", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
				Expect(serviceBindingRepo.DeleteApplicationGuid).To(Equal("my-app-guid"))
			})

			It("restage app after unbinds a service when force restage", func() {
				ui := callUnbindService([]string{"my-app", "my-service", "--force-restage"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)

				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding app", "my-service", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"Restaging app", "my-app", "my-org", "my-space", "my-user"},
				))
				Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
				Expect(serviceBindingRepo.DeleteApplicationGuid).To(Equal("my-app-guid"))
				Expect(stagingWatcher.watched).To(Equal(app))
				Expect(stagingWatcher.orgName).To(Equal(configRepo.OrganizationFields().Name))
				Expect(stagingWatcher.spaceName).To(Equal(configRepo.SpaceFields().Name))
			})

			It("do not restage app after unbinds a service when not force restage", func() {
				ui := callUnbindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)

				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding app", "my-service", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
				Expect(serviceBindingRepo.DeleteApplicationGuid).To(Equal("my-app-guid"))
				Expect(stagingWatcher.watched).To(Equal(models.Application{}))
				Expect(stagingWatcher.orgName).To(Equal(""))
				Expect(stagingWatcher.spaceName).To(Equal(""))
			})
		})

		Context("when the service instance does not exist", func() {
			BeforeEach(func() {
				serviceBindingRepo.DeleteBindingNotFound = true
			})

			It("warns the user the the service instance does not exist", func() {
				ui := callUnbindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)

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
			ui := callUnbindService([]string{"my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)
			Expect(ui.FailedWithUsage).To(BeTrue())

			ui = callUnbindService([]string{"my-app"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)
			Expect(ui.FailedWithUsage).To(BeTrue())

			ui = callUnbindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)
			Expect(ui.FailedWithUsage).To(BeFalse())
		})
	})
})

func callUnbindService(args []string, requirementsFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository, appRepo applications.ApplicationRepository, stagingWatcher ApplicationStagingWatcher) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUnbindService(fakeUI, config, serviceBindingRepo, appRepo, stagingWatcher)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}
