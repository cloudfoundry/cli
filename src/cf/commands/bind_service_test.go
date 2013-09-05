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

func TestBindCommand(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testhelpers.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceRepo := &testhelpers.FakeServiceRepo{}
	config := &configuration.Configuration{}
	fakeUI := callBindService([]string{"--service", "my-service", "--app", "my-app"}, config, reqFactory, serviceRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Binding service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")

	assert.Equal(t, serviceRepo.BindServiceServiceInstance, serviceInstance)
	assert.Equal(t, serviceRepo.BindServiceApplication, app)

	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestBindCommandIfServiceIsAlreadyBound(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testhelpers.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceRepo := &testhelpers.FakeServiceRepo{BindServiceErrorCode: 90003}
	config := &configuration.Configuration{}
	fakeUI := callBindService([]string{"--service", "my-service", "--app", "my-app"}, config, reqFactory, serviceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Binding service")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "is already bound")
}

func TestBindCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	serviceRepo := &testhelpers.FakeServiceRepo{}
	config := &configuration.Configuration{}

	fakeUI := callBindService([]string{"--service", "my-service"}, config, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callBindService([]string{"--app", "my-app"}, config, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callBindService([]string{"--app", "my-app", "--service", "my-service"}, config, reqFactory, serviceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func callBindService(args []string, config *configuration.Configuration, reqFactory *testhelpers.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("bind-service", args)
	cmd := NewBindService(fakeUI, config, serviceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
