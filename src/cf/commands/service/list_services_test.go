package service_test

import (
	"cf"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
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

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewListServices(ui, config, serviceSummaryRepo)
	cmd.Run(testcmd.NewContext("services", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting services in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
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
