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

func TestUpdateUserProvidedServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

	fakeUI := callUpdateUserProvidedService(t, []string{}, reqFactory, userProvidedServiceInstanceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUpdateUserProvidedService(t, []string{"foo"}, reqFactory, userProvidedServiceInstanceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUpdateUserProvidedService(t, []string{"foo", "bar"}, reqFactory, userProvidedServiceInstanceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestUpdateUserProvidedServiceRequirements(t *testing.T) {
	args := []string{"service-name", "values"}
	reqFactory := &testreq.FakeReqFactory{}
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

	reqFactory.LoginSuccess = false
	callUpdateUserProvidedService(t, args, reqFactory, userProvidedServiceInstanceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUpdateUserProvidedService(t, args, reqFactory, userProvidedServiceInstanceRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.ServiceInstanceName, "service-name")
}

func TestUpdateUserProvidedServiceWithJson(t *testing.T) {
	args := []string{"service-name", `{"foo":"bar"}`}
	serviceInstance := cf.ServiceInstance{Name: "found-service-name"}
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	expectedParams := map[string]string{"foo": "bar"}

	ui := callUpdateUserProvidedService(t, args, reqFactory, userProvidedServiceInstanceRepo)

	assert.Contains(t, ui.Outputs[0], "Updating user provided service")
	assert.Contains(t, ui.Outputs[0], "found-service-name")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, userProvidedServiceInstanceRepo.UpdateServiceInstance, serviceInstance)
	assert.Equal(t, userProvidedServiceInstanceRepo.UpdateParameters, expectedParams)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestUpdateUserProvidedServiceWithInvalidJson(t *testing.T) {
	args := []string{"service-name", `{"foo":"ba`}
	serviceInstance := cf.ServiceInstance{Name: "found-service-name"}
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

	ui := callUpdateUserProvidedService(t, args, reqFactory, userProvidedServiceInstanceRepo)

	assert.NotEqual(t, userProvidedServiceInstanceRepo.UpdateServiceInstance, serviceInstance)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "JSON is invalid")
}

func TestUpdateUserProvidedServiceWithAServiceInstanceThatIsNotUserProvided(t *testing.T) {
	args := []string{"service-name", `{"foo":"bar"}`}
	serviceInstance := cf.ServiceInstance{
		Name: "found-service-name",
		ServicePlan: cf.ServicePlan{
			Guid: "my-plan-guid",
		},
	}
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

	ui := callUpdateUserProvidedService(t, args, reqFactory, userProvidedServiceInstanceRepo)

	assert.NotEqual(t, userProvidedServiceInstanceRepo.UpdateServiceInstance, serviceInstance)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Service Instance is not user provided")
}

func callUpdateUserProvidedService(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("udpate-user-provided-service", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewUpdateUserProvidedService(fakeUI, config, userProvidedServiceInstanceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
