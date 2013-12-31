package service_test

import (
	"cf"
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

func TestRenameServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}

	fakeUI, _ := callRenameService(t, []string{}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI, _ = callRenameService(t, []string{"my-service"}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI, _ = callRenameService(t, []string{"my-service", "new-name", "extra"}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameServiceRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	callRenameService(t, []string{"my-service", "new-name"}, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	callRenameService(t, []string{"my-service", "new-name"}, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")
}

func TestRenameService(t *testing.T) {
	serviceInstance := cf.ServiceInstance{}
	serviceInstance.Name = "different-name"
	serviceInstance.Guid = "different-name-guid"
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstance: serviceInstance}
	ui, fakeServiceRepo := callRenameService(t, []string{"my-service", "new-name"}, reqFactory)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Renaming service", "different-name", "new-name", "my-org", "my-space", "my-user"},
		{"OK"},
	})

	assert.Equal(t, fakeServiceRepo.RenameServiceServiceInstance, serviceInstance)
	assert.Equal(t, fakeServiceRepo.RenameServiceNewName, "new-name")
}

func callRenameService(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, serviceRepo *testapi.FakeServiceRepo) {
	ui = &testterm.FakeUI{}
	serviceRepo = &testapi.FakeServiceRepo{}

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

	cmd := NewRenameService(ui, config, serviceRepo)
	ctxt := testcmd.NewContext("rename-service", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
