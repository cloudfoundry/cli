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
	testassert "testhelpers/assert"
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

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting services in org", "my-org", "my-space", "my-user"},
		{"OK"},
		{"my-service-1", "cleardb", "spark", "cli1, cli2"},
		{"my-service-2", "cleardb", "spark-2", "cli1"},
		{"my-service-provided-by-user", "user-provided"},
	})
}
