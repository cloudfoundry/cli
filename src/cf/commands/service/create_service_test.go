package service_test

import (
	"cf/api"
	. "cf/commands/service"
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

func callCreateService(args []string, inputs []string, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{Inputs: inputs}
	ctxt := testcmd.NewContext("create-service", args)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewCreateService(fakeUI, config, serviceRepo)
	reqFactory := &testreq.FakeReqFactory{}

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("create-service command", func() {
	It("successfully creates a service", func() {
		offering := models.ServiceOffering{}
		offering.Label = "cleardb"
		plan := models.ServicePlanFields{}
		plan.Name = "spark"
		plan.Guid = "cleardb-spark-guid"
		offering.Plans = []models.ServicePlanFields{plan}
		offering2 := models.ServiceOffering{}
		offering2.Label = "postgres"

		serviceRepo := &testapi.FakeServiceRepo{}
		serviceRepo.FindServiceOfferingsForSpaceByLabelReturns.ServiceOfferings = []models.ServiceOffering{
			offering,
			offering2,
		}

		ui := callCreateService([]string{"cleardb", "spark", "my-cleardb-service"},
			[]string{},
			serviceRepo,
		)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating service", "my-cleardb-service", "my-org", "my-space", "my-user"},
			{"OK"},
		})
		Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
	})

	It("warns the user when the service already exists with the same service plan", func() {
		offering := models.ServiceOffering{}
		offering.Label = "cleardb"
		plan := models.ServicePlanFields{}
		plan.Name = "spark"
		plan.Guid = "cleardb-spark-guid"
		offering.Plans = []models.ServicePlanFields{plan}
		offering2 := models.ServiceOffering{}
		offering2.Label = "postgres"

		serviceRepo := &testapi.FakeServiceRepo{}
		serviceRepo.CreateServiceInstanceReturns.Error = errors.NewServiceInstanceAlreadyExistsError("my-cleardb-service")
		serviceRepo.FindServiceOfferingsForSpaceByLabelReturns.ServiceOfferings = []models.ServiceOffering{offering, offering2}

		ui := callCreateService([]string{"cleardb", "spark", "my-cleardb-service"},
			[]string{},
			serviceRepo,
		)

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
			offering := models.ServiceOffering{}
			offering.Label = "cleardb"

			offering2 := models.ServiceOffering{}
			offering2.Label = "cleardb"

			plan := models.ServicePlanFields{}
			plan.Name = "spark"
			plan.Guid = "cleardb-spark-guid"
			offering2.Plans = []models.ServicePlanFields{plan}

			serviceRepo := &testapi.FakeServiceRepo{}
			serviceRepo.CreateServiceInstanceReturns.Error = errors.NewServiceInstanceAlreadyExistsError("my-cleardb-service")
			serviceRepo.FindServiceOfferingsForSpaceByLabelReturns.ServiceOfferings = []models.ServiceOffering{offering, offering2}
			ui := callCreateService([]string{"cleardb", "spark", "my-cleardb-service"},
				[]string{},
				serviceRepo,
			)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Creating service", "my-cleardb-service", "my-org", "my-space", "my-user"},
				{"OK"},
			})
			Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
			Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
		})
	})
})
