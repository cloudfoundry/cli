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

func TestUpdateUserProvidedServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

	ui := callUpdateUserProvidedService(t, []string{}, reqFactory, userProvidedServiceInstanceRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateUserProvidedService(t, []string{"foo"}, reqFactory, userProvidedServiceInstanceRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestUpdateUserProvidedServiceRequirements(t *testing.T) {
	args := []string{"service-name"}
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

func TestUpdateUserProvidedServiceWhenNoFlagsArePresent(t *testing.T) {
	args := []string{"service-name"}
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "found-service-name"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	ui := callUpdateUserProvidedService(t, args, reqFactory, repo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating user provided service", "found-service-name", "my-org", "my-space", "my-user"},
		{"OK"},
		{"No changes"},
	})
}

func TestUpdateUserProvidedServiceWithJson(t *testing.T) {
	args := []string{"-p", `{"foo":"bar"}`, "-l", "syslog://example.com", "service-name"}
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "found-service-name"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	ui := callUpdateUserProvidedService(t, args, reqFactory, repo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating user provided service", "found-service-name", "my-org", "my-space", "my-user"},
		{"OK"},
		{"TIP"},
	})
	assert.Equal(t, repo.UpdateServiceInstance.Name, serviceInstance.Name)
	assert.Equal(t, repo.UpdateServiceInstance.Params, map[string]string{"foo": "bar"})
	assert.Equal(t, repo.UpdateServiceInstance.SysLogDrainUrl, "syslog://example.com")
}

func TestUpdateUserProvidedServiceWithoutJson(t *testing.T) {
	args := []string{"-l", "syslog://example.com", "service-name"}
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "found-service-name"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	ui := callUpdateUserProvidedService(t, args, reqFactory, repo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating user provided service"},
		{"OK"},
	})
}

func TestUpdateUserProvidedServiceWithInvalidJson(t *testing.T) {
	args := []string{"-p", `{"foo":"ba`, "service-name"}
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "found-service-name"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

	ui := callUpdateUserProvidedService(t, args, reqFactory, userProvidedServiceInstanceRepo)

	assert.NotEqual(t, userProvidedServiceInstanceRepo.UpdateServiceInstance, serviceInstance)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"JSON is invalid"},
	})
}

func TestUpdateUserProvidedServiceWithAServiceInstanceThatIsNotUserProvided(t *testing.T) {
	args := []string{"-p", `{"foo":"bar"}`, "service-name"}
	plan := cf.ServicePlanFields{}
	plan.Guid = "my-plan-guid"
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "found-service-name"
	serviceInstance.ServicePlan = plan

	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:    true,
		ServiceInstance: serviceInstance,
	}
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

	ui := callUpdateUserProvidedService(t, args, reqFactory, userProvidedServiceInstanceRepo)

	assert.NotEqual(t, userProvidedServiceInstanceRepo.UpdateServiceInstance, serviceInstance)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Service Instance is not user provided"},
	})
}

func callUpdateUserProvidedService(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("update-user-provided-service", args)

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

	cmd := NewUpdateUserProvidedService(fakeUI, config, userProvidedServiceInstanceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
