package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestServices(t *testing.T) {
	serviceInstances := []cf.ServiceInstance{
		cf.ServiceInstance{
			Name: "my-service-1",
			ServicePlan: cf.ServicePlan{
				Name: "spark",
				ServiceOffering: cf.ServiceOffering{
					Label:    "cleardb",
					Provider: "cleardb provider",
					Version:  "1.0",
				},
			},
			ApplicationNames: []string{"cli1", "cli2"},
		},
		cf.ServiceInstance{
			Name: "my-service-2",
			ServicePlan: cf.ServicePlan{
				Name: "spark",
				ServiceOffering: cf.ServiceOffering{
					Label:    "cleardb",
					Provider: "cleardb provider",
					Version:  "1.1",
				},
			},
			ApplicationNames: []string{"cli1"},
		},
	}
	spaceRepo := &testhelpers.FakeSpaceRepository{SummarySpace: cf.Space{ServiceInstances: serviceInstances}}
	ui := &testhelpers.FakeUI{}
	config := &configuration.Configuration{
		Space: cf.Space{Name: "development", Guid: "development-guid"},
	}

	cmd := NewServices(ui, config, spaceRepo)
	cmd.Run(testhelpers.NewContext("services", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting services in development")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "my-service-1")
	assert.Contains(t, ui.Outputs[3], "cleardb")
	assert.Contains(t, ui.Outputs[3], "cleardb provider")
	assert.Contains(t, ui.Outputs[3], "1.0")
	assert.Contains(t, ui.Outputs[3], "spark")
	assert.Contains(t, ui.Outputs[3], "cli1, cli2")

	assert.Contains(t, ui.Outputs[4], "my-service-2")
	assert.Contains(t, ui.Outputs[4], "cleardb")
	assert.Contains(t, ui.Outputs[4], "cleardb provider")
	assert.Contains(t, ui.Outputs[4], "1.1")
	assert.Contains(t, ui.Outputs[4], "spark")
	assert.Contains(t, ui.Outputs[4], "cli1")
}
