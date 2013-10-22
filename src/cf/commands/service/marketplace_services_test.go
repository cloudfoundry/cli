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
	serviceOfferings := []cf.ServiceOffering{
		cf.ServiceOffering{
			Label:       "my-service-offering-1",
			Description: "service offering 1 description",
			Plans: []cf.ServicePlan{
				cf.ServicePlan{Name: "service-plan-a"},
				cf.ServicePlan{Name: "service-plan-b"},
			},
		},
		cf.ServiceOffering{
			Label:       "my-service-offering-2",
			Description: "service offering 2 description",
			Plans: []cf.ServicePlan{
				cf.ServicePlan{Name: "service-plan-c"},
				cf.ServicePlan{Name: "service-plan-d"},
			},
		},
	}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space", Guid: "my-space-guid"},
		Organization: cf.Organization{Name: "my-org", Guid: "my-org-guid"},
		AccessToken:  token,
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
