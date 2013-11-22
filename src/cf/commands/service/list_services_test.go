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
	plan := cf.ServicePlanFields{}
	plan.Guid = "spark-guid"
	plan.Name = "spark"

	offering := cf.ServiceOfferingFields{}
	offering.Label = "cleardb"

	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "my-service-1"
	serviceInstance.ServicePlan = plan
	serviceInstance.ApplicationNames = []string{"cli1", "cli2"}
	serviceInstance.ServiceOffering = offering

	plan2 := cf.ServicePlanFields{}
	plan2.Guid = "spark-guid-2"
	plan2.Name = "spark-2"

	serviceInstance2 := cf.ServiceInstance{}
	serviceInstance2.Name = "my-service-2"
	serviceInstance2.ServicePlan = plan2
	serviceInstance2.ApplicationNames = []string{"cli1"}
	serviceInstance2.ServiceOffering = offering

	serviceInstance3 := cf.ServiceInstance{}
	serviceInstance3.Name = "my-service-provided-by-user"

	serviceInstances := []cf.ServiceInstance{serviceInstance, serviceInstance2, serviceInstance3}
	serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
		GetSummariesInCurrentSpaceInstances: serviceInstances,
	}
	ui := &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
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
	assert.Contains(t, ui.Outputs[5], "spark-2")
	assert.Contains(t, ui.Outputs[5], "cli1")

	assert.Contains(t, ui.Outputs[6], "my-service-provided-by-user")
	assert.Contains(t, ui.Outputs[6], "user-provided")
}
