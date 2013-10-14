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

func TestSetOrgRoleFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSetOrgRole([]string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callSetOrgRole([]string{"my-user", "my-org"}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetOrgRole([]string{"my-user"}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetOrgRole([]string{}, reqFactory, userRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestSetOrgRoleRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	userRepo := &testapi.FakeUserRepository{}

	reqFactory.LoginSuccess = false
	callSetOrgRole([]string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callSetOrgRole([]string{"my-user", "my-org", "my-role"}, reqFactory, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "my-user")
	assert.Equal(t, reqFactory.OrganizationName, "my-org")
}

func TestSetOrgRole(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		User:         cf.User{Guid: "my-user-guid", Username: "my-user"},
		Organization: cf.Organization{Guid: "my-org-guid", Name: "my-org"},
	}
	userRepo := &testapi.FakeUserRepository{}

	ui := callSetOrgRole([]string{"some-user", "some-org", "some-role"}, reqFactory, userRepo)

	assert.Contains(t, ui.Outputs[0], "Assigning")
	assert.Contains(t, ui.Outputs[0], "some-role")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[0], "my-org")

	assert.Equal(t, userRepo.SetOrgRoleUser, reqFactory.User)
	assert.Equal(t, userRepo.SetOrgRoleOrganization, reqFactory.Organization)
	assert.Equal(t, userRepo.SetOrgRoleRole, "some-role")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callSetOrgRole(args []string, reqFactory *testreq.FakeReqFactory, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-user", args)
	cmd := NewSetOrgRole(ui, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
