package service_test

import (
	. "cf/commands/service"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("services", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.Repository
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
				testcmd.RunCommand(cmd, testcmd.NewContext("services", []string{}), requirementsFactory)
				Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
			})
		})

		Context("when no space is targeted", func() {
			BeforeEach(func() {
				requirementsFactory.TargetedSpaceSuccess = false
			})

			It("fails requirements", func() {
				testcmd.RunCommand(cmd, testcmd.NewContext("services", []string{}), requirementsFactory)
				Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
			})
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
		serviceInstance.ServicePlan = plan
		serviceInstance.ApplicationNames = []string{"cli1", "cli2"}
		serviceInstance.ServiceOffering = offering

		serviceInstance2 := models.ServiceInstance{}
		serviceInstance2.Name = "my-service-2"
		serviceInstance2.ServicePlan = plan2
		serviceInstance2.ApplicationNames = []string{"cli1"}
		serviceInstance2.ServiceOffering = offering

		serviceInstance3 := models.ServiceInstance{}
		serviceInstance3.Name = "my-service-provided-by-user"

		serviceInstances := []models.ServiceInstance{serviceInstance, serviceInstance2, serviceInstance3}
		serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
			GetSummariesInCurrentSpaceInstances: serviceInstances,
		}

		cmd := NewListServices(ui, configRepo, serviceSummaryRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("services", []string{}), requirementsFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting services in org", "my-org", "my-space", "my-user"},
			{"OK"},
			{"my-service-1", "cleardb", "spark", "cli1, cli2"},
			{"my-service-2", "cleardb", "spark-2", "cli1"},
			{"my-service-provided-by-user", "user-provided"},
		})
	})

	It("lists no services when none are found", func() {
		serviceInstances := []models.ServiceInstance{}
		serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
			GetSummariesInCurrentSpaceInstances: serviceInstances,
		}

		cmd := NewListServices(ui, configRepo, serviceSummaryRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("services", []string{}), requirementsFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting services in org", "my-org", "my-space", "my-user"},
			{"OK"},
			{"No services found"},
		})
		testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
			{"name", "service", "plan", "bound apps"},
		})
	})
})
