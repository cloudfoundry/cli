package service_test

import (
	"cf"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testterm "testhelpers/terminal"
	"testing"
)

func TestServices(t *testing.T) {
	serviceInstances := []cf.ServiceInstance{
		cf.ServiceInstance{
			Name: "my-service-1",
			ServicePlan: cf.ServicePlan{
				Guid: "spark-guid",
				Name: "spark",
				ServiceOffering: cf.ServiceOffering{
					Label: "cleardb",
				},
			},
			ApplicationNames: []string{"cli1", "cli2"},
		},
		cf.ServiceInstance{
			Name: "my-service-2",
			ServicePlan: cf.ServicePlan{
				Guid: "spark-guid",
				Name: "spark",
				ServiceOffering: cf.ServiceOffering{
					Label: "cleardb",
				},
			},
			ApplicationNames: []string{"cli1"},
		},
		cf.ServiceInstance{
			Name: "my-service-provided-by-user",
		},
	}
	serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
		GetSummariesInCurrentSpaceInstances: serviceInstances,
	}
	ui := &testterm.FakeUI{}
	config := &configuration.Configuration{
		Space: cf.Space{Name: "development"},
	}

	cmd := NewListServices(ui, config, serviceSummaryRepo)
	cmd.Run(testcmd.NewContext("services", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting services in")
	assert.Contains(t, ui.Outputs[0], "development")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[4], "my-service-1")
	assert.Contains(t, ui.Outputs[4], "cleardb")
	assert.Contains(t, ui.Outputs[4], "spark")
	assert.Contains(t, ui.Outputs[4], "cli1, cli2")

	assert.Contains(t, ui.Outputs[5], "my-service-2")
	assert.Contains(t, ui.Outputs[5], "cleardb")
	assert.Contains(t, ui.Outputs[5], "spark")
	assert.Contains(t, ui.Outputs[5], "cli1")

	assert.Contains(t, ui.Outputs[6], "my-service-provided-by-user")
	assert.Contains(t, ui.Outputs[6], "user-provided")
}
