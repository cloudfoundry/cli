package service_test

import (
	. "cf/commands/service"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestServices", func() {

			plan := models.ServicePlanFields{}
			plan.Guid = "spark-guid"
			plan.Name = "spark"

			offering := models.ServiceOfferingFields{}
			offering.Label = "cleardb"

			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service-1"
			serviceInstance.ServicePlan = plan
			serviceInstance.ApplicationNames = []string{"cli1", "cli2"}
			serviceInstance.ServiceOffering = offering

			plan2 := models.ServicePlanFields{}
			plan2.Guid = "spark-guid-2"
			plan2.Name = "spark-2"

			serviceInstance2 := models.ServiceInstance{}
			serviceInstance2.Name = "my-service-2"
			serviceInstance2.ServicePlan = plan2
			serviceInstance2.ApplicationNames = []string{"cli1"}
			serviceInstance2.ServiceOffering = offering

			serviceInstance3 := models.ServiceInstance{}
			serviceInstance3.Name = "my-service-provided-by-user"

			serviceInstances := []models.ServiceInstance{serviceInstance, serviceInstance2, serviceInstance3}
			serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
				GetSummariesInCurrentSpaceInstances: serviceInstances,
			}
			ui := &testterm.FakeUI{}

			token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
				Username: "my-user",
			})
			assert.NoError(mr.T(), err)
			org := models.OrganizationFields{}
			org.Name = "my-org"
			space := models.SpaceFields{}
			space.Name = "my-space"
			config := &configuration.Configuration{
				SpaceFields:        space,
				OrganizationFields: org,
				AccessToken:        token,
			}

			cmd := NewListServices(ui, config, serviceSummaryRepo)
			cmd.Run(testcmd.NewContext("services", []string{}))

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting services in org", "my-org", "my-space", "my-user"},
				{"OK"},
				{"my-service-1", "cleardb", "spark", "cli1, cli2"},
				{"my-service-2", "cleardb", "spark-2", "cli1"},
				{"my-service-provided-by-user", "user-provided"},
			})
		})
		It("TestEmptyServicesList", func() {

			serviceInstances := []models.ServiceInstance{}
			serviceSummaryRepo := &testapi.FakeServiceSummaryRepo{
				GetSummariesInCurrentSpaceInstances: serviceInstances,
			}
			ui := &testterm.FakeUI{}

			token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
				Username: "my-user",
			})
			assert.NoError(mr.T(), err)
			org := models.OrganizationFields{}
			org.Name = "my-org"
			space := models.SpaceFields{}
			space.Name = "my-space"
			config := &configuration.Configuration{
				SpaceFields:        space,
				OrganizationFields: org,
				AccessToken:        token,
			}

			cmd := NewListServices(ui, config, serviceSummaryRepo)
			cmd.Run(testcmd.NewContext("services", []string{}))

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting services in org", "my-org", "my-space", "my-user"},
				{"OK"},
				{"No services found"},
			})
			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"name", "service", "plan", "bound apps"},
			})
		})
	})
}
