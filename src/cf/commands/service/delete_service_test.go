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

func TestDeleteServiceCommandWithY(t *testing.T) {
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "my-service"
	serviceInstance.Guid = "my-service-guid"
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	fakeUI := callDeleteService(t, "Y", []string{"my-service"}, reqFactory, serviceRepo)

	assert.Contains(t, fakeUI.Prompts[0], "Are you sure")

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")

	assert.Equal(t, serviceRepo.DeleteServiceServiceInstance, serviceInstance)
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestDeleteServiceCommandWithYes(t *testing.T) {
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "my-service"
	serviceInstance.Guid = "my-service-guid"
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	fakeUI := callDeleteService(t, "Yes", []string{"my-service"}, reqFactory, serviceRepo)

	assert.Contains(t, fakeUI.Prompts[0], "Are you sure")

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")

	assert.Equal(t, serviceRepo.DeleteServiceServiceInstance, serviceInstance)
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestDeleteServiceCommandOnNonExistentService(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
	fakeUI := callDeleteService(t, "", []string{"-f", "my-service"}, reqFactory, serviceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")

	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-service")
	assert.Contains(t, fakeUI.Outputs[2], "not exist")
}

func TestDeleteServiceCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{}

	fakeUI := callDeleteService(t, "", []string{"-f"}, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callDeleteService(t, "", []string{"-f", "my-service"}, reqFactory, serviceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestDeleteServiceForceFlagSkipsConfirmation(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{}

	ui := callDeleteService(t, "", []string{"-f", "foo.com"}, reqFactory, serviceRepo)

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting service")
	assert.Contains(t, ui.Outputs[0], "foo.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callDeleteService(t *testing.T, confirmation string, args []string, reqFactory *testreq.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-service", args)

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

	cmd := NewDeleteService(fakeUI, config, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
