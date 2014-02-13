package service_test

import (
	"cf/api"
	. "cf/commands/service"
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

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateService", func() {
		offering := models.ServiceOffering{}
		offering.Label = "cleardb"
		plan := models.ServicePlanFields{}
		plan.Name = "spark"
		plan.Guid = "cleardb-spark-guid"
		offering.Plans = []models.ServicePlanFields{plan}
		offering2 := models.ServiceOffering{}
		offering2.Label = "postgres"

		serviceRepo := &testapi.FakeServiceRepo{}
		serviceRepo.GetAllServiceOfferingsReturns.ServiceOfferings = []models.ServiceOffering{
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
		Expect(serviceRepo.CreateServiceInstanceName).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstancePlanGuid).To(Equal("cleardb-spark-guid"))
	})

	It("TestCreateServiceWhenServiceAlreadyExists", func() {
		offering := models.ServiceOffering{}
		offering.Label = "cleardb"
		plan := models.ServicePlanFields{}
		plan.Name = "spark"
		plan.Guid = "cleardb-spark-guid"
		offering.Plans = []models.ServicePlanFields{plan}
		offering2 := models.ServiceOffering{}
		offering2.Label = "postgres"
		serviceRepo := &testapi.FakeServiceRepo{CreateServiceAlreadyExists: true}
		serviceRepo.GetAllServiceOfferingsReturns.ServiceOfferings = []models.ServiceOffering{offering, offering2}
		ui := callCreateService([]string{"cleardb", "spark", "my-cleardb-service"},
			[]string{},
			serviceRepo,
		)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating service", "my-cleardb-service"},
			{"OK"},
			{"my-cleardb-service", "already exists"},
		})
		Expect(serviceRepo.CreateServiceInstanceName).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstancePlanGuid).To(Equal("cleardb-spark-guid"))
	})
})
