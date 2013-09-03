package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateCommand(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{
		cf.ServiceOffering{Label: "cleardb", Plans: []cf.ServicePlan{
			cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"},
		}},
		cf.ServiceOffering{Label: "postgres"},
	}
	serviceRepo := &testhelpers.FakeServiceRepo{ServiceOfferings: serviceOfferings}
	config := &configuration.Configuration{}
	fakeUI := callCreateService([]string{"--offering", "cleardb", "--plan", "spark", "--name", "my-cleardb-service"}, config, serviceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating service")
	assert.Contains(t, fakeUI.Outputs[0], "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstanceName, "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstancePlan, cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"})
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callCreateService(args []string, config *configuration.Configuration, serviceRepo api.ServiceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	target := NewCreateService(fakeUI, config, serviceRepo)
	target.Run(testhelpers.NewContext("create-service", args))
	return
}
