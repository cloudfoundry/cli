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

func TestUnbindCommand(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testreq.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{}
	fakeUI := callUnbindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Unbinding service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")

	assert.Equal(t, serviceBindingRepo.DeleteServiceInstance, serviceInstance)
	assert.Equal(t, serviceBindingRepo.DeleteApplication, app)

	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestUnbindCommandWhenBindingIsNonExistent(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testreq.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{DeleteBindingNotFound: true}
	fakeUI := callUnbindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Unbinding service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")

	assert.Equal(t, serviceBindingRepo.DeleteServiceInstance, serviceInstance)
	assert.Equal(t, serviceBindingRepo.DeleteApplication, app)

	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-service")
	assert.Contains(t, fakeUI.Outputs[2], "my-app")
	assert.Contains(t, fakeUI.Outputs[2], "did not exist")
}

func TestUnbindCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

	fakeUI := callUnbindService([]string{"my-service"}, reqFactory, serviceBindingRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUnbindService([]string{"my-app"}, reqFactory, serviceBindingRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUnbindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func callUnbindService(args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("unbind-service", args)
	cmd := NewUnbindService(fakeUI, serviceBindingRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
