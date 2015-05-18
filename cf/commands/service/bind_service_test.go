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

var _ = Describe("bind-service command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		stagingWatcher      *fakeStagingWatcher
		appRepo             *testApplication.FakeApplicationRepository
		configRepo          core_config.ReadWriter
		serviceRepo         *testapi.FakeServiceRepo
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{}
		appRepo = &testApplication.FakeApplicationRepository{}
		stagingWatcher = &fakeStagingWatcher{}
		serviceRepo = &testapi.FakeServiceRepo{}
	})

	It("fails requirements when not logged in", func() {
		cmd := NewBindService(&testterm.FakeUI{}, testconfig.NewRepository(), &testapi.FakeServiceBindingRepo{}, appRepo, stagingWatcher)

		Expect(testcmd.RunCommand(cmd, []string{"service", "app"}, requirementsFactory)).To(BeFalse())
	})

	It("fails requirements when service not found", func() {
		serviceRepo.FindInstanceByNameNotFound = true

		cmd := NewBindService(&testterm.FakeUI{}, testconfig.NewRepository(), &testapi.FakeServiceBindingRepo{}, appRepo, stagingWatcher)

		Expect(testcmd.RunCommand(cmd, []string{"service", "app"}, requirementsFactory)).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			configRepo = testconfig.NewRepositoryWithDefaults()
		})

		It("binds a service instance to an app", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"
			requirementsFactory.Application = app
			requirementsFactory.ServiceInstance = serviceInstance
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{}
			ui := callBindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"TIP"},
			))
			Expect(serviceBindingRepo.CreateServiceInstanceGuid).To(Equal("my-service-guid"))
			Expect(serviceBindingRepo.CreateApplicationGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when the service instance is already bound to the given app", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"
			requirementsFactory.Application = app
			requirementsFactory.ServiceInstance = serviceInstance
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{CreateErrorCode: "90003"}
			ui := callBindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding service"},
				[]string{"OK"},
				[]string{"my-app", "is already bound", "my-service"},
			))
		})

		It("warns the user when the error is non HttpError ", func() {
			app := models.Application{}
			app.Name = "my-app1"
			app.Guid = "my-app1-guid1"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service1"
			serviceInstance.Guid = "my-service1-guid1"
			requirementsFactory.Application = app
			requirementsFactory.ServiceInstance = serviceInstance
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{CreateNonHttpErrCode: "1001"}
			ui := callBindService([]string{"my-app1", "my-service1"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				[]string{"FAILED"},
				[]string{"1001"},
			))
		})

		It("fails with usage when called without a service instance and app", func() {
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

			ui := callBindService([]string{"my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)
			Expect(ui.FailedWithUsage).To(BeTrue())

			ui = callBindService([]string{"my-app"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)
			Expect(ui.FailedWithUsage).To(BeTrue())

			ui = callBindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)
			Expect(ui.FailedWithUsage).To(BeFalse())
		})

		It("restage app after bind when force restage", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"
			requirementsFactory.Application = app
			requirementsFactory.ServiceInstance = serviceInstance
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{}
			ui := callBindService([]string{"my-app", "my-service", "--force-restage"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"Restaging app", "my-app", "my-org", "my-space", "my-user"},
			))
			Expect(serviceBindingRepo.CreateServiceInstanceGuid).To(Equal("my-service-guid"))
			Expect(serviceBindingRepo.CreateApplicationGuid).To(Equal("my-app-guid"))
			Expect(stagingWatcher.watched).To(Equal(app))
			Expect(stagingWatcher.orgName).To(Equal(configRepo.OrganizationFields().Name))
			Expect(stagingWatcher.spaceName).To(Equal(configRepo.SpaceFields().Name))
		})

		It("do not restage app after bind when not force restage", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"
			requirementsFactory.Application = app
			requirementsFactory.ServiceInstance = serviceInstance
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{}
			ui := callBindService([]string{"my-app", "my-service"}, requirementsFactory, serviceBindingRepo, appRepo, stagingWatcher)

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))
			Expect(serviceBindingRepo.CreateServiceInstanceGuid).To(Equal("my-service-guid"))
			Expect(serviceBindingRepo.CreateApplicationGuid).To(Equal("my-app-guid"))
			Expect(stagingWatcher.watched).To(Equal(models.Application{}))
			Expect(stagingWatcher.orgName).To(Equal(""))
			Expect(stagingWatcher.spaceName).To(Equal(""))
		})
	})
})

func callBindService(args []string, requirementsFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository, appRepo applications.ApplicationRepository, stagingWatcher ApplicationStagingWatcher) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewBindService(fakeUI, config, serviceBindingRepo, appRepo, stagingWatcher)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

type fakeStagingWatcher struct {
	watched   models.Application
	orgName   string
	spaceName string
}

func (f *fakeStagingWatcher) ApplicationWatchStaging(app models.Application, orgName, spaceName string, start func(models.Application) (models.Application, error)) (updatedApp models.Application, err error) {
	f.watched = app
	f.orgName = orgName
	f.spaceName = spaceName
	return start(app)
}
