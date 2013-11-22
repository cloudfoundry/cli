package service_test

import (
	"cf"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestMarketplaceServices(t *testing.T) {
	plan := cf.ServicePlanFields{}
	plan.Name = "service-plan-a"
	plan2 := cf.ServicePlanFields{}
	plan2.Name = "service-plan-b"
	plan3 := cf.ServicePlanFields{}
	plan3.Name = "service-plan-c"
	plan4 := cf.ServicePlanFields{}
	plan4.Name = "service-plan-d"

	offering := cf.ServiceOffering{}
	offering.Label = "my-service-offering-1"
	offering.Description = "service offering 1 description"
	offering.Plans = []cf.ServicePlanFields{plan, plan2}

	offering2 := cf.ServiceOffering{}
	offering2.Label = "my-service-offering-2"
	offering2.Description = "service offering 2 description"
	offering2.Plans = []cf.ServicePlanFields{plan3, plan4}

	serviceOfferings := []cf.ServiceOffering{offering, offering2}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
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

	ui := callMarketplaceServices(t, config, serviceRepo)

	assert.Contains(t, ui.Outputs[0], "Getting services from marketplace in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[4], "my-service-offering-1")
	assert.Contains(t, ui.Outputs[4], "service offering 1 description")
	assert.Contains(t, ui.Outputs[4], "service-plan-a, service-plan-b")

	assert.Contains(t, ui.Outputs[5], "my-service-offering-2")
	assert.Contains(t, ui.Outputs[5], "service offering 2 description")
	assert.Contains(t, ui.Outputs[5], "service-plan-c, service-plan-d")
}

func TestMarketplaceServicesWhenNotLoggedIn(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

	config := &configuration.Configuration{}

	ui := callMarketplaceServices(t, config, serviceRepo)

	assert.Contains(t, ui.Outputs[0], "Getting services from marketplace...")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callMarketplaceServices(t *testing.T, config *configuration.Configuration, serviceRepo *testapi.FakeServiceRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("marketplace", []string{})
	reqFactory := &testreq.FakeReqFactory{}

	cmd := NewMarketplaceServices(ui, config, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
