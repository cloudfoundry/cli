package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestUnbindCommand(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testhelpers.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceRepo := &testhelpers.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	fakeUI := callUnbindService([]string{"--service", "my-service", "--app", "my-app"}, reqFactory, serviceRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Unbinding service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")

	assert.Equal(t, serviceRepo.UnbindServiceServiceInstance, serviceInstance)
	assert.Equal(t, serviceRepo.UnbindServiceApplication, app)

	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestUnbindCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	serviceRepo := &testhelpers.FakeServiceRepo{}

	fakeUI := callUnbindService([]string{"--service", "my-service"}, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUnbindService([]string{"--app", "my-app"}, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUnbindService([]string{"--app", "my-app", "--service", "my-service"}, reqFactory, serviceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func callUnbindService(args []string, reqFactory *testhelpers.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("unbind-service", args)
	cmd := NewUnbindService(fakeUI, serviceRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
