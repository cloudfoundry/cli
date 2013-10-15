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

func TestUnsetOrgRoleFailsWithUsage(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callUnsetOrgRole([]string{}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetOrgRole([]string{"username"}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetOrgRole([]string{"username", "org"}, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetOrgRole([]string{"username", "org", "role"}, userRepo, reqFactory)
	assert.False(t, ui.FailedWithUsage)
}

func TestUnsetOrgRoleRequirements(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"username", "org", "role"}

	reqFactory.LoginSuccess = false
	callUnsetOrgRole(args, userRepo, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUnsetOrgRole(args, userRepo, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "username")
	assert.Equal(t, reqFactory.OrganizationName, "org")
}

func TestUnsetOrgRole(t *testing.T) {
	userRepo := &testapi.FakeUserRepository{}

	user := cf.User{Username: "some-user", Guid: "some-user-guid"}
	org := cf.Organization{Name: "some-org", Guid: "some-org-guid"}
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		User:         user,
		Organization: org,
	}
	args := []string{"my-username", "my-org", "my-role"}

	ui := callUnsetOrgRole(args, userRepo, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Removing ")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-username")
	assert.Contains(t, ui.Outputs[0], "my-role")

	assert.Equal(t, userRepo.UnsetOrgRoleRole, "my-role")
	assert.Equal(t, userRepo.UnsetOrgRoleUser, user)
	assert.Equal(t, userRepo.UnsetOrgRoleOrganization, org)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callUnsetOrgRole(args []string, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("unset-org-role", args)
	cmd := NewUnsetOrgRole(ui, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
