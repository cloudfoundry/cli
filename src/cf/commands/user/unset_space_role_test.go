package user_test

import (
	"cf"
	. "cf/commands/user"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestUnsetSpaceRoleFailsWithUsage(t *testing.T) {
	reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()

	ui := callUnsetSpaceRole(t, []string{}, spaceRepo, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole(t, []string{"username"}, spaceRepo, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole(t, []string{"username", "org"}, spaceRepo, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole(t, []string{"username", "org", "space"}, spaceRepo, userRepo, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnsetSpaceRole(t, []string{"username", "org", "space", "role"}, spaceRepo, userRepo, reqFactory)
	assert.False(t, ui.FailedWithUsage)
}

func TestUnsetSpaceRoleRequirements(t *testing.T) {
	reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
	args := []string{"username", "org", "space", "role"}

	reqFactory.LoginSuccess = false
	callUnsetSpaceRole(t, args, spaceRepo, userRepo, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUnsetSpaceRole(t, args, spaceRepo, userRepo, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "username")
	assert.Equal(t, reqFactory.OrganizationName, "org")
}

func TestUnsetSpaceRole(t *testing.T) {
	user := cf.UserFields{}
	user.Username = "some-user"
	user.Guid = "some-user-guid"
	org := cf.Organization{}
	org.Name = "some-org"
	org.Guid = "some-org-guid"

	reqFactory, spaceRepo, userRepo := getUnsetSpaceRoleDeps()
	reqFactory.LoginSuccess = true
	reqFactory.UserFields = user
	reqFactory.Organization = org
	spaceRepo.FindByNameInOrgSpace = cf.Space{}
	spaceRepo.FindByNameInOrgSpace.Name = "some-space"
	spaceRepo.FindByNameInOrgSpace.Guid = "some-space-guid"

	args := []string{"my-username", "my-org", "my-space", "SpaceManager"}

	ui := callUnsetSpaceRole(t, args, spaceRepo, userRepo, reqFactory)

	assert.Equal(t, spaceRepo.FindByNameInOrgName, "my-space")
	assert.Equal(t, spaceRepo.FindByNameInOrgOrgGuid, "some-org-guid")

	assert.Contains(t, ui.Outputs[0], "Removing role ")
	assert.Contains(t, ui.Outputs[0], "SpaceManager")
	assert.Contains(t, ui.Outputs[0], "some-user")
	assert.Contains(t, ui.Outputs[0], "some-org")
	assert.Contains(t, ui.Outputs[0], "some-space")
	assert.Contains(t, ui.Outputs[0], "current-user")

	assert.Equal(t, userRepo.UnsetSpaceRoleRole, cf.SPACE_MANAGER)
	assert.Equal(t, userRepo.UnsetSpaceRoleUserGuid, "some-user-guid")
	assert.Equal(t, userRepo.UnsetSpaceRoleSpaceGuid, "some-space-guid")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func getUnsetSpaceRoleDeps() (reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	reqFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}

func callUnsetSpaceRole(t *testing.T, args []string, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("unset-space-role", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	space2 := cf.SpaceFields{}
	space2.Name = "my-space"
	org2 := cf.OrganizationFields{}
	org2.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space2,
		OrganizationFields: org2,
		AccessToken:        token,
	}

	cmd := NewUnsetSpaceRole(ui, config, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
