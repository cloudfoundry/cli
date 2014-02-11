package service_test

import (
	. "cf/commands/service"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callMarketplaceServices(t mr.TestingT, config configuration.Reader, serviceRepo *testapi.FakeServiceRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("marketplace", []string{})
	reqFactory := &testreq.FakeReqFactory{}

	cmd := NewMarketplaceServices(ui, config, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing Marketplace Services", func() {

	var config configuration.ReadWriter

	Context("when the user is logged in", func() {

		BeforeEach(func() {
			config = testconfig.NewRepositoryWithDefaults()
		})

		It("lists the correct service offerings", func() {
			plan := models.ServicePlanFields{}
			plan.Name = "service-plan-a"
			plan2 := models.ServicePlanFields{}
			plan2.Name = "service-plan-b"
			plan3 := models.ServicePlanFields{}
			plan3.Name = "service-plan-c"
			plan4 := models.ServicePlanFields{}
			plan4.Name = "service-plan-d"

			offering := models.ServiceOffering{}
			offering.Label = "zzz-my-service-offering"
			offering.Description = "service offering 1 description"
			offering.Plans = []models.ServicePlanFields{plan, plan2}

			offering2 := models.ServiceOffering{}
			offering2.Label = "aaa-my-service-offering"
			offering2.Description = "service offering 2 description"
			offering2.Plans = []models.ServicePlanFields{plan3, plan4}

			serviceOfferings := []models.ServiceOffering{offering, offering2}
			serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

			ui := callMarketplaceServices(mr.T(), config, serviceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting services from marketplace in org", "my-org", "my-space", "my-user"},
				{"OK"},
				{"service", "plans", "description"},
				{"aaa-my-service-offering", "service offering 2 description", "service-plan-c", "service-plan-d"},
				{"zzz-my-service-offering", "service offering 1 description", "service-plan-a", "service-plan-b"},
			})
		})

	})

	Context("when user is not logged in", func() {

		BeforeEach(func() {
			config = testconfig.NewRepository()
		})

		It("fails gracefully when user is not logged in", func() {
			serviceOfferings := []models.ServiceOffering{}
			serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

			ui := callMarketplaceServices(mr.T(), config, serviceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting services from marketplace..."},
				{"OK"},
				{"No service offerings found"},
			})
			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"service", "plans", "description"},
			})
		})

	})
})
