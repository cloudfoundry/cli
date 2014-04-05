package service_test

import (
	. "cf/commands/service"
	"cf/configuration"
	"cf/errors"
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

var _ = Describe("create-service command", func() {
	var (
		ui          *testterm.FakeUI
		config      configuration.Repository
		reqFactory  *testreq.FakeReqFactory
		cmd         CreateService
		serviceRepo *testapi.FakeServiceRepo

		offering1 models.ServiceOffering
		offering2 models.ServiceOffering
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		serviceRepo = &testapi.FakeServiceRepo{}
		cmd = NewCreateService(ui, config, serviceRepo)

		offering1 = models.ServiceOffering{}
		offering1.Label = "cleardb"
		offering1.Plans = []models.ServicePlanFields{{
			Name: "spark",
			Guid: "cleardb-spark-guid",
		}}

		offering2 = models.ServiceOffering{}
		offering2.Label = "postgres"

		serviceRepo.FindServiceOfferingsForSpaceByLabelReturns.ServiceOfferings = []models.ServiceOffering{offering1, offering2}
	})

	var callCreateService = func(args []string) {
		ctxt := testcmd.NewContext("create-service", args)
		testcmd.RunCommand(cmd, ctxt, reqFactory)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails when not logged in", func() {
			reqFactory.LoginSuccess = false
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			reqFactory.TargetedSpaceSuccess = false
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("successfully creates a service", func() {
		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating service", "my-cleardb-service", "my-org", "my-space", "my-user"},
			{"OK"},
		})
		Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
	})

	It("warns the user when the service already exists with the same service plan", func() {
		serviceRepo.CreateServiceInstanceReturns.Error = errors.NewModelAlreadyExistsError("Service", "my-cleardb-service")

		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating service", "my-cleardb-service"},
			{"OK"},
			{"my-cleardb-service", "already exists"},
		})
		Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
	})

	Context("When there are multiple services with the same label", func() {
		It("finds the plan even if it has to search multiple services", func() {
			offering2.Label = "cleardb"

			serviceRepo.CreateServiceInstanceReturns.Error = errors.NewModelAlreadyExistsError("Service", "my-cleardb-service")
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating service", "my-cleardb-service", "my-org", "my-space", "my-user"},
				{"OK"},
			})
			Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
			Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
		})
	})
})
