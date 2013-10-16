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

func TestBindCommand(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testreq.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{}
	fakeUI := callBindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Binding service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")

	assert.Equal(t, serviceBindingRepo.CreateServiceInstance, serviceInstance)
	assert.Equal(t, serviceBindingRepo.CreateApplication, app)

	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "TIP")
	assert.Equal(t, len(fakeUI.Outputs), 3)
}

func TestBindCommandIfServiceIsAlreadyBound(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testreq.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{CreateErrorCode: "90003"}
	fakeUI := callBindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, len(fakeUI.Outputs), 3)
	assert.Contains(t, fakeUI.Outputs[0], "Binding service")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-app")
	assert.Contains(t, fakeUI.Outputs[2], "is already bound")
	assert.Contains(t, fakeUI.Outputs[2], "my-service")
}

func TestBindCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

	fakeUI := callBindService([]string{"my-service"}, reqFactory, serviceBindingRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callBindService([]string{"my-app"}, reqFactory, serviceBindingRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callBindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func callBindService(args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("bind-service", args)
	cmd := NewBindService(fakeUI, serviceBindingRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
