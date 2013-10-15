package user_test

import (
	"cf"
	. "cf/commands/user"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestUnsetSpaceRoleFailsWithUsage(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callUnsetSpaceRole([]string{}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole([]string{"username"}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole([]string{"username", "space"}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole([]string{"username", "space", "role"}, userRepo, reqFactory)
	assert.False(t, ui.FailedWithUsage)
}

func TestUnsetSpaceRoleRequirements(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"username", "space", "role"}

	reqFactory.LoginSuccess = false
	callUnsetSpaceRole(args, userRepo, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUnsetSpaceRole(args, userRepo, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "username")
	assert.Equal(t, reqFactory.SpaceName, "space")
}

func TestUnsetSpaceRole(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}

	user := cf.User{Username: "some-user", Guid: "some-user-guid"}
	space := cf.Space{Name: "some-space", Guid: "some-space-guid"}
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		User:         user,
		Space:        space,
	}
	args := []string{"my-username", "my-space", "my-role"}

	ui := callUnsetSpaceRole(args, userRepo, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Removing")
	assert.Contains(t, ui.Outputs[0], "some-space")
	assert.Contains(t, ui.Outputs[0], "some-user")
	assert.Contains(t, ui.Outputs[0], "my-role")

	assert.Equal(t, userRepo.UnsetSpaceRoleRole, "my-role")
	assert.Equal(t, userRepo.UnsetSpaceRoleUser, user)
	assert.Equal(t, userRepo.UnsetSpaceRoleSpace, space)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callUnsetSpaceRole(args []string, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("unset-space-role", args)
	cmd := NewUnsetSpaceRole(ui, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
