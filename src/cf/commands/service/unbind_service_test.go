package service_test

import (
	"cf"
	"cf/api"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestUnbindCommand(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "my-service"
	serviceInstance.Guid = "my-service-guid"
	reqFactory := &testreq.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{}
	ui := callUnbindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Unbinding app", "my-service", "my-app", "my-org", "my-space", "my-user"},
		{"OK"},
	})
	assert.Equal(t, serviceBindingRepo.DeleteServiceInstance, serviceInstance)
	assert.Equal(t, serviceBindingRepo.DeleteApplicationGuid, "my-app-guid")
}

func TestUnbindCommandWhenBindingIsNonExistent(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "my-service"
	serviceInstance.Guid = "my-service-guid"
	reqFactory := &testreq.FakeReqFactory{
		Application:     app,
		ServiceInstance: serviceInstance,
	}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{DeleteBindingNotFound: true}
	ui := callUnbindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Unbinding app", "my-service", "my-app"},
		{"OK"},
		{"my-service", "my-app", "did not exist"},
	})
	assert.Equal(t, serviceBindingRepo.DeleteServiceInstance, serviceInstance)
	assert.Equal(t, serviceBindingRepo.DeleteApplicationGuid, "my-app-guid")
}

func TestUnbindCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

	ui := callUnbindService(t, []string{"my-service"}, reqFactory, serviceBindingRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnbindService(t, []string{"my-app"}, reqFactory, serviceBindingRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnbindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
	assert.False(t, ui.FailedWithUsage)
}

func callUnbindService(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("unbind-service", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewUnbindService(fakeUI, config, serviceBindingRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
