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

func TestDeleteServiceCommandWithY(t *testing.T) {
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "my-service"
	serviceInstance.Guid = "my-service-guid"
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	ui := callDeleteService(t, "Y", []string{"my-service"}, reqFactory, serviceRepo)

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Are you sure"},
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service", "my-service", "my-org", "my-space", "my-user"},
		{"OK"},
	})

	assert.Equal(t, serviceRepo.DeleteServiceServiceInstance, serviceInstance)
}

func TestDeleteServiceCommandWithYes(t *testing.T) {
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "my-service"
	serviceInstance.Guid = "my-service-guid"
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	ui := callDeleteService(t, "Yes", []string{"my-service"}, reqFactory, serviceRepo)

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{{"Are you sure"}})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service", "my-service"},
		{"OK"},
	})

	assert.Equal(t, serviceRepo.DeleteServiceServiceInstance, serviceInstance)
}

func TestDeleteServiceCommandOnNonExistentService(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
	ui := callDeleteService(t, "", []string{"-f", "my-service"}, reqFactory, serviceRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service", "my-service"},
		{"OK"},
		{"my-service", "does not exist"},
	})
}

func TestDeleteServiceCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{}

	ui := callDeleteService(t, "", []string{"-f"}, reqFactory, serviceRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteService(t, "", []string{"-f", "my-service"}, reqFactory, serviceRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteServiceForceFlagSkipsConfirmation(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{}

	ui := callDeleteService(t, "", []string{"-f", "foo.com"}, reqFactory, serviceRepo)

	assert.Equal(t, len(ui.Prompts), 0)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service", "foo.com"},
		{"OK"},
	})
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
