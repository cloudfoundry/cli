package service_test

import (
	"cf"
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
	serviceInstance := cf.ServiceInstance{Name: "different-name", Guid: "different-name-guid"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstance: serviceInstance}
	fakeUI, fakeServiceRepo := callRenameService(t, []string{"my-service", "new-name"}, reqFactory)

	assert.Contains(t, fakeUI.Outputs[0], "Renaming service")
	assert.Contains(t, fakeUI.Outputs[0], "different-name")
	assert.Contains(t, fakeUI.Outputs[0], "new-name")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Equal(t, fakeUI.Outputs[1], "OK")

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

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewRenameService(ui, config, serviceRepo)
	ctxt := testcmd.NewContext("rename-service", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
