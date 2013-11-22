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
	fakeUI := callBindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")

	assert.Contains(t, fakeUI.Outputs[0], "Binding service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")

	assert.Equal(t, serviceBindingRepo.CreateServiceInstanceGuid, "my-service-guid")
	assert.Equal(t, serviceBindingRepo.CreateApplicationGuid, "my-app-guid")

	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "TIP")
	assert.Equal(t, len(fakeUI.Outputs), 3)
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
	fakeUI := callBindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

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

	fakeUI := callBindService(t, []string{"my-service"}, reqFactory, serviceBindingRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callBindService(t, []string{"my-app"}, reqFactory, serviceBindingRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callBindService(t, []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
	assert.False(t, fakeUI.FailedWithUsage)
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
