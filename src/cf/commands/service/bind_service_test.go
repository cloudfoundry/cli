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

func TestBindCommand(t *testing.T) {
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
	ui := callBindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Equal(t, len(ui.Outputs), 3)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
		{"OK"},
		{"TIP"},
	})
	assert.Equal(t, serviceBindingRepo.CreateServiceInstanceGuid, "my-service-guid")
	assert.Equal(t, serviceBindingRepo.CreateApplicationGuid, "my-app-guid")
}

func TestBindCommandIfServiceIsAlreadyBound(t *testing.T) {
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
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{CreateErrorCode: "90003"}
	ui := callBindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Binding service"},
		{"OK"},
		{"my-app", "is already bound", "my-service"},
	})
}

func TestBindCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

	ui := callBindService(t, []string{"my-service"}, reqFactory, serviceBindingRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callBindService(t, []string{"my-app"}, reqFactory, serviceBindingRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callBindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
	assert.False(t, ui.FailedWithUsage)
}

func callBindService(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("bind-service", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewBindService(fakeUI, config, serviceBindingRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
