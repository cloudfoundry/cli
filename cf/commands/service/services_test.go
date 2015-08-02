package service_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/plugin/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("services", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		serviceSummaryRepo  *testapi.FakeServiceSummaryRepo
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetServiceSummaryRepository(serviceSummaryRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("services").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("services", args, requirementsFactory, updateCommandDependency, false)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceSummaryRepo = &testapi.FakeServiceSummaryRepo{}
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
			TargetedOrgSuccess:   true,
		}

		deps = command_registry.NewDependency()
	})

	Describe("services requirements", func() {

		Context("when not logged in", func() {
			BeforeEach(func() {
				requirementsFactory.LoginSuccess = false
			})

			It("fails requirements", func() {
				Expect(runCommand()).To(BeFalse())
			})
		})

		Context("when no space is targeted", func() {
			BeforeEach(func() {
				requirementsFactory.TargetedSpaceSuccess = false
			})

			It("fails requirements", func() {
				Expect(runCommand()).To(BeFalse())
			})
		})
		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			Expect(runCommand("blahblah")).To(BeFalse())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "No argument required"},
			))
		})
	})

	It("lists available services", func() {
		plan := models.ServicePlanFields{
			Guid: "spark-guid",
			Name: "spark",
		}

		plan2 := models.ServicePlanFields{
			Guid: "spark-guid-2",
			Name: "spark-2",
		}

		offering := models.ServiceOfferingFields{Label: "cleardb"}

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "my-service-1"
		serviceInstance.LastOperation.Type = "create"
		serviceInstance.LastOperation.State = "in progress"
		serviceInstance.LastOperation.Description = "fake state description"
		serviceInstance.ServicePlan = plan
		serviceInstance.ApplicationNames = []string{"cli1", "cli2"}
		serviceInstance.ServiceOffering = offering

		serviceInstance2 := models.ServiceInstance{}
		serviceInstance2.Name = "my-service-2"
		serviceInstance2.LastOperation.Type = "create"
		serviceInstance2.LastOperation.State = ""
		serviceInstance2.LastOperation.Description = "fake state description"
		serviceInstance2.ServicePlan = plan2
		serviceInstance2.ApplicationNames = []string{"cli1"}
		serviceInstance2.ServiceOffering = offering

		userProvidedServiceInstance := models.ServiceInstance{}
		userProvidedServiceInstance.Name = "my-service-provided-by-user"

		serviceInstances := []models.ServiceInstance{serviceInstance, serviceInstance2, userProvidedServiceInstance}

		serviceSummaryRepo.GetSummariesInCurrentSpaceInstances = serviceInstances

		runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting services in org", "my-org", "my-space", "my-user"},
			[]string{"name", "service", "plan", "bound apps", "last operation"},
			[]string{"OK"},
			[]string{"my-service-1", "cleardb", "spark", "cli1, cli2", "create in progress"},
			[]string{"my-service-2", "cleardb", "spark-2", "cli1", ""},
			[]string{"my-service-provided-by-user", "user-provided", "", "", ""},
		))
	})

	It("lists no services when none are found", func() {
		serviceInstances := []models.ServiceInstance{}
		serviceSummaryRepo.GetSummariesInCurrentSpaceInstances = serviceInstances

		runCommand()

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting services in org", "my-org", "my-space", "my-user"},
			[]string{"OK"},
			[]string{"No services found"},
		))

		Expect(ui.Outputs).ToNot(ContainSubstrings(
			[]string{"name", "service", "plan", "bound apps"},
		))
	})

	Describe("when invoked by a plugin", func() {

		var (
			pluginModels []plugin_models.GetServices_Model
		)

		BeforeEach(func() {

			pluginModels = []plugin_models.GetServices_Model{}
			deps.PluginModels.Services = &pluginModels
			plan := models.ServicePlanFields{
				Guid: "spark-guid",
				Name: "spark",
			}

			plan2 := models.ServicePlanFields{
				Guid: "spark-guid-2",
				Name: "spark-2",
			}

			offering := models.ServiceOfferingFields{Label: "cleardb"}

			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service-1"
			serviceInstance.Guid = "123"
			serviceInstance.LastOperation.Type = "create"
			serviceInstance.LastOperation.State = "in progress"
			serviceInstance.LastOperation.Description = "fake state description"
			serviceInstance.ServicePlan = plan
			serviceInstance.ApplicationNames = []string{"cli1", "cli2"}
			serviceInstance.ServiceOffering = offering

			serviceInstance2 := models.ServiceInstance{}
			serviceInstance2.Name = "my-service-2"
			serviceInstance2.Guid = "345"
			serviceInstance2.LastOperation.Type = "create"
			serviceInstance2.LastOperation.State = ""
			serviceInstance2.LastOperation.Description = "fake state description"
			serviceInstance2.ServicePlan = plan2
			serviceInstance2.ApplicationNames = []string{"cli1"}
			serviceInstance2.ServiceOffering = offering

			userProvidedServiceInstance := models.ServiceInstance{}
			userProvidedServiceInstance.Name = "my-service-provided-by-user"
			userProvidedServiceInstance.Guid = "678"

			serviceInstances := []models.ServiceInstance{serviceInstance, serviceInstance2, userProvidedServiceInstance}

			serviceSummaryRepo.GetSummariesInCurrentSpaceInstances = serviceInstances
		})

		It("populates the plugin model", func() {
			testcmd.RunCliCommand("services", []string{}, requirementsFactory, updateCommandDependency, true)

			Ω(len(pluginModels)).To(Equal(3))
			Ω(pluginModels[0].Name).To(Equal("my-service-1"))
			Ω(pluginModels[0].Guid).To(Equal("123"))
			Ω(pluginModels[0].ServicePlan.Name).To(Equal("spark"))
			Ω(pluginModels[0].ServicePlan.Guid).To(Equal("spark-guid"))
			Ω(pluginModels[0].Service.Name).To(Equal("cleardb"))
			Ω(pluginModels[0].ApplicationNames).To(Equal([]string{"cli1", "cli2"}))
			Ω(pluginModels[0].LastOperation.Type).To(Equal("create"))
			Ω(pluginModels[0].LastOperation.State).To(Equal("in progress"))
			Ω(pluginModels[0].IsUserProvided).To(BeFalse())

			Ω(pluginModels[1].Name).To(Equal("my-service-2"))
			Ω(pluginModels[1].Guid).To(Equal("345"))
			Ω(pluginModels[1].ServicePlan.Name).To(Equal("spark-2"))
			Ω(pluginModels[1].ServicePlan.Guid).To(Equal("spark-guid-2"))
			Ω(pluginModels[1].Service.Name).To(Equal("cleardb"))
			Ω(pluginModels[1].ApplicationNames).To(Equal([]string{"cli1"}))
			Ω(pluginModels[1].LastOperation.Type).To(Equal("create"))
			Ω(pluginModels[1].LastOperation.State).To(Equal(""))
			Ω(pluginModels[1].IsUserProvided).To(BeFalse())

			Ω(pluginModels[2].Name).To(Equal("my-service-provided-by-user"))
			Ω(pluginModels[2].Guid).To(Equal("678"))
			Ω(pluginModels[2].ServicePlan.Name).To(Equal(""))
			Ω(pluginModels[2].ServicePlan.Guid).To(Equal(""))
			Ω(pluginModels[2].Service.Name).To(Equal(""))
			Ω(pluginModels[2].ApplicationNames).To(BeNil())
			Ω(pluginModels[2].LastOperation.Type).To(Equal(""))
			Ω(pluginModels[2].LastOperation.State).To(Equal(""))
			Ω(pluginModels[2].IsUserProvided).To(BeTrue())

		})

	})
})
