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

func TestUpdateUserProvidedServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{}

	fakeUI := callUpdateUserProvidedService([]string{}, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUpdateUserProvidedService([]string{"foo"}, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUpdateUserProvidedService([]string{"foo", "bar"}, reqFactory, serviceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestUpdateUserProvidedServiceRequirements(t *testing.T) {
	args := []string{"service-name", "values"}
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{}

	reqFactory.LoginSuccess = false
	callUpdateUserProvidedService(args, reqFactory, serviceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUpdateUserProvidedService(args, reqFactory, serviceRepo)
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
	serviceRepo := &testapi.FakeServiceRepo{}
	expectedParams := map[string]string{"foo": "bar"}

	ui := callUpdateUserProvidedService(args, reqFactory, serviceRepo)

	assert.Contains(t, ui.Outputs[0], "Updating user provided service")
	assert.Contains(t, ui.Outputs[0], "found-service-name")

	assert.Equal(t, serviceRepo.UpdateUserProvidedServiceInstanceServiceInstance, serviceInstance)
	assert.Equal(t, serviceRepo.UpdateUserProvidedServiceInstanceParameters, expectedParams)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestUpdateUserProvidedServiceWithInvalidJson(t *testing.T) {
	args := []string{"service-name", `{"foo":"ba`}
	serviceInstance := cf.ServiceInstance{Name: "found-service-name"}
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	serviceRepo := &testapi.FakeServiceRepo{}

	ui := callUpdateUserProvidedService(args, reqFactory, serviceRepo)

	assert.NotEqual(t, serviceRepo.UpdateUserProvidedServiceInstanceServiceInstance, serviceInstance)

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
	serviceRepo := &testapi.FakeServiceRepo{}

	ui := callUpdateUserProvidedService(args, reqFactory, serviceRepo)

	assert.NotEqual(t, serviceRepo.UpdateUserProvidedServiceInstanceServiceInstance, serviceInstance)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Service Instance is not user provided")
}

func callUpdateUserProvidedService(args []string, reqFactory *testreq.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("udpate-user-provided-service", args)
	cmd := NewUpdateUserProvidedService(fakeUI, serviceRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
