package service_test

import (
	"cf"
	"cf/api"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
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
	fakeUI := callUnbindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Unbinding app")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")

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
	fakeUI := callUnbindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Unbinding app")
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

	fakeUI := callUnbindService(t, []string{"my-service"}, reqFactory, serviceBindingRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUnbindService(t, []string{"my-app"}, reqFactory, serviceBindingRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUnbindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func callUnbindService(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("unbind-service", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewUnbindService(fakeUI, config, serviceBindingRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
