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

func TestSetSpaceRoleFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSetSpaceRole([]string{}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole([]string{"my-user"}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole([]string{"my-user", "my-space"}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole([]string{"my-user", "my-space", "my-role"}, reqFactory, userRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestSetSpaceRoleRequirements(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"username", "space", "role"}

	reqFactory.LoginSuccess = false
	callSetSpaceRole(args, reqFactory, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callSetSpaceRole(args, reqFactory, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "username")
	assert.Equal(t, reqFactory.SpaceName, "space")
}

func TestSetSpaceRole(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		User:         cf.User{Guid: "my-user-guid", Username: "my-user"},
		Space:        cf.Space{Guid: "my-space-guid", Name: "my-space"},
	}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSetSpaceRole([]string{"some-user", "some-space", "some-role"}, reqFactory, userRepo)

	assert.Contains(t, ui.Outputs[0], "Assigning")
	assert.Contains(t, ui.Outputs[0], "some-role")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[0], "my-space")

	assert.Equal(t, userRepo.SetSpaceRoleUser, reqFactory.User)
	assert.Equal(t, userRepo.SetSpaceRoleSpace, reqFactory.Space)
	assert.Equal(t, userRepo.SetSpaceRoleRole, "some-role")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callSetSpaceRole(args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-space-role", args)
	cmd := NewSetSpaceRole(ui, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
