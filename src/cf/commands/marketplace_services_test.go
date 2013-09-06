package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestMarketplaceServices(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{
		cf.ServiceOffering{
			Label:       "my-service-offering-1",
			Provider:    "my-service-offering-1 provider",
			Description: "service offering 1 description",
			Version:     "1.0",
			Plans: []cf.ServicePlan{
				cf.ServicePlan{Name: "service-plan-a"},
				cf.ServicePlan{Name: "service-plan-b"},
			},
		},
		cf.ServiceOffering{
			Label:       "my-service-offering-2",
			Provider:    "my-service-offering-2 provider",
			Description: "service offering 2 description",
			Version:     "1.4",
			Plans: []cf.ServicePlan{
				cf.ServicePlan{Name: "service-plan-c"},
				cf.ServicePlan{Name: "service-plan-d"},
			},
		},
	}
	serviceRepo := &testhelpers.FakeServiceRepo{ServiceOfferings: serviceOfferings}
	ui := &testhelpers.FakeUI{}

	ctxt := testhelpers.NewContext("services", []string{"--marketplace"})
	reqFactory := &testhelpers.FakeReqFactory{}

	cmd := NewMarketplaceServices(ui, serviceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Getting services from marketplace...")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "my-service-offering-1")
	assert.Contains(t, ui.Outputs[3], "my-service-offering-1 provider")
	assert.Contains(t, ui.Outputs[3], "service offering 1 description")
	assert.Contains(t, ui.Outputs[3], "1.0")
	assert.Contains(t, ui.Outputs[3], "service-plan-a, service-plan-b")

	assert.Contains(t, ui.Outputs[4], "my-service-offering-2")
	assert.Contains(t, ui.Outputs[4], "my-service-offering-2 provider")
	assert.Contains(t, ui.Outputs[4], "service offering 2 description")
	assert.Contains(t, ui.Outputs[4], "1.4")
	assert.Contains(t, ui.Outputs[4], "service-plan-c, service-plan-d")
}
