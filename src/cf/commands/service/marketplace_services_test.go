package service_test

import (
	"cf"
	. "cf/commands/service"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callMarketplaceServices(t mr.TestingT, config *configuration.Configuration, serviceRepo *testapi.FakeServiceRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("marketplace", []string{})
	reqFactory := &testreq.FakeReqFactory{}

	cmd := NewMarketplaceServices(ui, config, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestMarketplaceServices", func() {
			plan := cf.ServicePlanFields{}
			plan.Name = "service-plan-a"
			plan2 := cf.ServicePlanFields{}
			plan2.Name = "service-plan-b"
			plan3 := cf.ServicePlanFields{}
			plan3.Name = "service-plan-c"
			plan4 := cf.ServicePlanFields{}
			plan4.Name = "service-plan-d"

			offering := cf.ServiceOffering{}
			offering.Label = "zzz-my-service-offering"
			offering.Description = "service offering 1 description"
			offering.Plans = []cf.ServicePlanFields{plan, plan2}

			offering2 := cf.ServiceOffering{}
			offering2.Label = "aaa-my-service-offering"
			offering2.Description = "service offering 2 description"
			offering2.Plans = []cf.ServicePlanFields{plan3, plan4}

			serviceOfferings := []cf.ServiceOffering{offering, offering2}
			serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

			token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
				Username: "my-user",
			})
			assert.NoError(mr.T(), err)
			org := cf.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			space := cf.SpaceFields{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"
			config := &configuration.Configuration{
				SpaceFields:        space,
				OrganizationFields: org,
				AccessToken:        token,
			}

			ui := callMarketplaceServices(mr.T(), config, serviceRepo)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting services from marketplace in org", "my-org", "my-space", "my-user"},
				{"OK"},
				{"service", "plans", "description"},
				{"aaa-my-service-offering", "service offering 2 description", "service-plan-c", "service-plan-d"},
				{"zzz-my-service-offering", "service offering 1 description", "service-plan-a", "service-plan-b"},
			})
		})
		It("TestMarketplaceServicesWhenNotLoggedIn", func() {

			serviceOfferings := []cf.ServiceOffering{}
			serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

			config := &configuration.Configuration{}

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
}
