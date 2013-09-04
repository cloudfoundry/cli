package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDeleteServiceCommand(t *testing.T) {
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testhelpers.FakeReqFactory{ServiceInstance: serviceInstance}
	serviceRepo := &testhelpers.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	config := &configuration.Configuration{}
	fakeUI := callDeleteService([]string{"my-service"}, config, reqFactory, serviceRepo)

	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")

	assert.Equal(t, serviceRepo.DeleteServiceServiceInstance, serviceInstance)
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestDeleteServiceCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	serviceRepo := &testhelpers.FakeServiceRepo{}
	config := &configuration.Configuration{}

	fakeUI := callDeleteService([]string{}, config, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callDeleteService([]string{"my-service"}, config, reqFactory, serviceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func callDeleteService(args []string, config *configuration.Configuration, reqFactory requirements.Factory, serviceRepo api.ServiceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("delete-service", args)
	cmd := NewDeleteService(fakeUI, config, serviceRepo)
	cmd.GetRequirements(reqFactory, ctxt)
	cmd.Run(ctxt)

	return
}
