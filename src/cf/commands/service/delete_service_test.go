package service_test

import (
	"cf"
	"cf/api"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteServiceCommand(t *testing.T) {
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	fakeUI := callDeleteService([]string{"my-service"}, reqFactory, serviceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")

	assert.Equal(t, serviceRepo.DeleteServiceServiceInstance, serviceInstance)
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestDeleteServiceCommandOnNonExistentService(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
	fakeUI := callDeleteService([]string{"my-service"}, reqFactory, serviceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")

	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-service")
	assert.Contains(t, fakeUI.Outputs[2], "not exist")
}

func TestDeleteServiceCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{}

	fakeUI := callDeleteService([]string{}, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callDeleteService([]string{"my-service"}, reqFactory, serviceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func callDeleteService(args []string, reqFactory *testreq.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("delete-service", args)
	cmd := NewDeleteService(fakeUI, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
