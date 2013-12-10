package user_test

import (
	"cf"
	. "cf/commands/user"
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

func TestSetSpaceRoleFailsWithUsage(t *testing.T) {
	reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

	ui := callSetSpaceRole(t, []string{}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole(t, []string{"my-user"}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole(t, []string{"my-user", "my-org"}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole(t, []string{"my-user", "my-org", "my-space"}, reqFactory, spaceRepo, userRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callSetSpaceRole(t, []string{"my-user", "my-org", "my-space", "my-role"}, reqFactory, spaceRepo, userRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestSetSpaceRoleRequirements(t *testing.T) {
	args := []string{"username", "org", "space", "role"}
	reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()

	reqFactory.LoginSuccess = false
	callSetSpaceRole(t, args, reqFactory, spaceRepo, userRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callSetSpaceRole(t, args, reqFactory, spaceRepo, userRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.UserUsername, "username")
	assert.Equal(t, reqFactory.OrganizationName, "org")
}

func TestSetSpaceRole(t *testing.T) {
	org := cf.Organization{}
	org.Guid = "my-org-guid"
	org.Name = "my-org"

	args := []string{"some-user", "some-org", "some-space", "SpaceManager"}

	reqFactory, spaceRepo, userRepo := getSetSpaceRoleDeps()
	reqFactory.LoginSuccess = true
	reqFactory.UserFields = cf.UserFields{}
	reqFactory.UserFields.Guid = "my-user-guid"
	reqFactory.UserFields.Username = "my-user"
	reqFactory.Organization = org

	spaceRepo.FindByNameInOrgSpace = cf.Space{}
	spaceRepo.FindByNameInOrgSpace.Guid = "my-space-guid"
	spaceRepo.FindByNameInOrgSpace.Name = "my-space"
	spaceRepo.FindByNameInOrgSpace.Organization = org.OrganizationFields

	ui := callSetSpaceRole(t, args, reqFactory, spaceRepo, userRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Assigning role ", "SpaceManager", "my-user", "my-org", "my-space", "current-user"},
		{"OK"},
	})

	assert.Equal(t, spaceRepo.FindByNameInOrgName, "some-space")
	assert.Equal(t, spaceRepo.FindByNameInOrgOrgGuid, "my-org-guid")

	assert.Equal(t, userRepo.SetSpaceRoleUserGuid, "my-user-guid")
	assert.Equal(t, userRepo.SetSpaceRoleSpaceGuid, "my-space-guid")
	assert.Equal(t, userRepo.SetSpaceRoleRole, cf.SPACE_MANAGER)
}

func getSetSpaceRoleDeps() (reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) {
	reqFactory = &testreq.FakeReqFactory{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	userRepo = &testapi.FakeUserRepository{}
	return
}

func callSetSpaceRole(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-space-role", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "current-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewSetSpaceRole(ui, config, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
