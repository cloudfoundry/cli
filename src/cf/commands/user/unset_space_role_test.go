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
	reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()

	ui := callUnsetSpaceRole([]string{}, spaceRepo, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole([]string{"username"}, spaceRepo, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole([]string{"username", "org"}, spaceRepo, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole([]string{"username", "org", "space"}, spaceRepo, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole([]string{"username", "org", "space", "role"}, spaceRepo, userRepo, reqFactory)
	assert.False(t, ui.FailedWithUsage)
}

func TestUnsetSpaceRoleRequirements(t *testing.T) {
	reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
	args := []string{"username", "org", "space", "role"}

	reqFactory.LoginSuccess = false
	callUnsetSpaceRole(args, spaceRepo, userRepo, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUnsetSpaceRole(args, spaceRepo, userRepo, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "username")
	assert.Equal(t, reqFactory.OrganizationName, "org")
}

func TestUnsetSpaceRole(t *testing.T) {
	user := cf.User{Username: "some-user", Guid: "some-user-guid"}
	org := cf.Organization{Name: "some-org", Guid: "some-org-guid"}

	reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
	reqFactory.LoginSuccess = true
	reqFactory.User = user
	reqFactory.Organization = org

	spaceRepo.FindByNameInOrgSpace = cf.Space{Name: "some-space"}

	args := []string{"my-username", "my-org", "my-space", "my-role"}

	ui := callUnsetSpaceRole(args, spaceRepo, userRepo, reqFactory)

	assert.Equal(t, spaceRepo.FindByNameInOrgName, "my-space")
	assert.Equal(t, spaceRepo.FindByNameInOrgOrg, reqFactory.Organization)

	assert.Contains(t, ui.Outputs[0], "Removing")
	assert.Contains(t, ui.Outputs[0], "some-org")
	assert.Contains(t, ui.Outputs[0], "some-space")
	assert.Contains(t, ui.Outputs[0], "some-user")
	assert.Contains(t, ui.Outputs[0], "my-role")

	assert.Equal(t, userRepo.UnsetSpaceRoleRole, "my-role")
	assert.Equal(t, userRepo.UnsetSpaceRoleUser, user)
	assert.Equal(t, userRepo.UnsetSpaceRoleSpace, spaceRepo.FindByNameInOrgSpace)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func getUnsetSpaceRoleDeps() (reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	reqFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}

func callUnsetSpaceRole(args []string, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("unset-space-role", args)
	cmd := NewUnsetSpaceRole(ui, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
