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

func TestDeleteServiceCommand(t *testing.T) {
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}

	serviceRepo := &testhelpers.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	config := &configuration.Configuration{}
	fakeUI := callDeleteService([]string{"my-service"}, config, serviceRepo)

	assert.Equal(t, serviceRepo.FindInstanceByNameName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")

	assert.Equal(t, serviceRepo.DeleteServiceServiceInstance, serviceInstance)
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callDeleteService(args []string, config *configuration.Configuration, serviceRepo api.ServiceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	target := NewDeleteService(fakeUI, config, serviceRepo)
	target.Run(testhelpers.NewContext("unbind-service", args))
	return
}
