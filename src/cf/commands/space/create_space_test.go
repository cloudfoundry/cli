package space_test

import (
	"cf"
	. "cf/commands/space"
	"cf/commands/user"
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

var (
	defaultOrg       cf.OrganizationFields
	defaultSpace     cf.Space
	configSpace      cf.SpaceFields
	defaultSpaceRepo *testapi.FakeSpaceRepository
	defaultUserRepo  *testapi.FakeUserRepository
)

func resetSpaceDefaults() {
	defaultOrg = cf.OrganizationFields{}
	defaultOrg.Name = "my-org"
	defaultOrg.Guid = "my-org-guid"

	defaultSpace = cf.Space{}
	defaultSpace.Name = "my-space"
	defaultSpace.Guid = "my-space-guid"
	defaultSpace.Organization = defaultOrg

	configSpace = cf.SpaceFields{}
	configSpace.Name = "config-space"
	configSpace.Guid = "config-space-guid"

	defaultSpaceRepo = &testapi.FakeSpaceRepository{
		CreateSpaceSpace: defaultSpace,
	}
	defaultUserRepo = &testapi.FakeUserRepository{}
}

func TestCreateSpaceFailsWithUsage(t *testing.T) {
	resetSpaceDefaults()
	reqFactory := &testreq.FakeReqFactory{}

	ui := callCreateSpace(t, []string{}, reqFactory, defaultSpaceRepo, defaultUserRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestCreateSpaceRequirements(t *testing.T) {
	resetSpaceDefaults()
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

}

func TestCreateSpace(t *testing.T) {
	resetSpaceDefaults()
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	ui := callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating space", "my-space", "my-org", "my-user"},
		{"OK"},
		{"Assigning", "my-user", "my-space", cf.SpaceRoleToUserInput[cf.SPACE_MANAGER]},
		{"Assigning", "my-user", "my-space", cf.SpaceRoleToUserInput[cf.SPACE_DEVELOPER]},
		{"TIP"},
	})

	assert.Equal(t, defaultSpaceRepo.CreateSpaceName, "my-space")
	assert.Equal(t, defaultSpaceRepo.CreateSpaceOrgGuid, "my-org-guid")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleUserGuid, "my-user-guid")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleSpaceGuid, "my-space-guid")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleRole, cf.SPACE_DEVELOPER)
}

func TestCreateSpaceWhenItAlreadyExists(t *testing.T) {
	resetSpaceDefaults()
	defaultSpaceRepo.CreateSpaceExists = true

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	ui := callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating space", "my-space"},
		{"OK"},
		{"my-space", "already exists"},
	})
	testassert.SliceDoesNotContain(t, ui.Outputs, testassert.Lines{
		{"Assigning", "my-user", "my-space", cf.SpaceRoleToUserInput[cf.SPACE_MANAGER]},
	})

	assert.Equal(t, defaultSpaceRepo.CreateSpaceName, "")
	assert.Equal(t, defaultSpaceRepo.CreateSpaceOrgGuid, "")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleUserGuid, "")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleSpaceGuid, "")
}

func callCreateSpace(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
		UserGuid: "my-user-guid",
	})
	assert.NoError(t, err)

	org := cf.OrganizationFields{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	config := &configuration.Configuration{
		SpaceFields:        configSpace,
		OrganizationFields: org,
		AccessToken:        token,
	}

	spaceRoleSetter := user.NewSetSpaceRole(ui, config, spaceRepo, userRepo)
	cmd := NewCreateSpace(ui, config, spaceRoleSetter, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
