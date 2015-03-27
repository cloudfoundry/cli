package service_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/cloudfoundry/cli/cf/commands/service"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("services", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
			TargetedOrgSuccess:   true,
		}
	})

	Describe("services requirements", func() {
		var cmd ListServices

		BeforeEach(func() {
			cmd = NewListServices(ui, configRepo, &testapi.FakeServiceSummaryRepo{})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				requirementsFactory.LoginSuccess = false
			})

			It("fails requirements", func() {
				Expect(testcmd.RunCommand(cmd, []string{}, requirementsFactory)).To(BeFalse())
			})
		})

		Context("when no space is targeted", func() {
			BeforeEach(func() {
				requirementsFactory.TargetedSpaceSuccess = false
			})

			It("fails requirements", func() {
				Expect(testcmd.RunCommand(cmd, []string{}, requirementsFactory)).To(BeFalse())
			})
		})
		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			Expect(testcmd.RunCommand(cmd, []string{"blahblah"}, requirementsFactory)).To(BeFalse())
			Expect(ui.FailedWithUsage).To(BeTrue())
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
		serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
			GetSummariesInCurrentSpaceInstances: serviceInstances,
		}

		cmd := NewListServices(ui, configRepo, serviceSummaryRepo)
		testcmd.RunCommand(cmd, []string{}, requirementsFactory)

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
		serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
			GetSummariesInCurrentSpaceInstances: serviceInstances,
		}

		cmd := NewListServices(ui, configRepo, serviceSummaryRepo)
		testcmd.RunCommand(cmd, []string{}, requirementsFactory)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting services in org", "my-org", "my-space", "my-user"},
			[]string{"OK"},
			[]string{"No services found"},
		))

		Expect(ui.Outputs).ToNot(ContainSubstrings(
			[]string{"name", "service", "plan", "bound apps"},
		))
	})
})
