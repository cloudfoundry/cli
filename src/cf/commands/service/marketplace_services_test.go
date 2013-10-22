package service_test

import (
	"cf"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
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
	ui := &testterm.FakeUI{}

	ctxt := testcmd.NewContext("marketplace", []string{})
	reqFactory := &testreq.FakeReqFactory{}

	cmd := NewMarketplaceServices(ui, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Getting services from marketplace...")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[4], "my-service-offering-1")
	assert.Contains(t, ui.Outputs[4], "service offering 1 description")
	assert.Contains(t, ui.Outputs[4], "service-plan-a, service-plan-b")

	assert.Contains(t, ui.Outputs[5], "my-service-offering-2")
	assert.Contains(t, ui.Outputs[5], "service offering 2 description")
	assert.Contains(t, ui.Outputs[5], "service-plan-c, service-plan-d")
}
