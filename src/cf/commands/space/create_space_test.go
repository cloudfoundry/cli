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
	"testhelpers/maker"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

var (
	defaultReqFactory *testreq.FakeReqFactory
	configSpace       cf.SpaceFields
	configOrg         cf.OrganizationFields
	defaultSpace      cf.Space
	defaultSpaceRepo  *testapi.FakeSpaceRepository
	defaultOrgRepo    *testapi.FakeOrgRepository
	defaultUserRepo   *testapi.FakeUserRepository
)

func resetSpaceDefaults() {
	defaultReqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	configOrg = cf.OrganizationFields{}
	configOrg.Name = "my-org"
	configOrg.Guid = "my-org-guid"

	configSpace = cf.SpaceFields{}
	configSpace.Name = "config-space"
	configSpace.Guid = "config-space-guid"

	defaultSpace = cf.Space{}
	defaultSpace.Name = "my-space"
	defaultSpace.Guid = "my-space-guid"
	defaultSpace.Organization = configOrg

	defaultSpaceRepo = &testapi.FakeSpaceRepository{
		CreateSpaceSpace: defaultSpace,
	}

	defaultUserRepo = &testapi.FakeUserRepository{}
	defaultOrgRepo = &testapi.FakeOrgRepository{}
}

func TestCreateSpaceFailsWithUsage(t *testing.T) {
	resetSpaceDefaults()
	reqFactory := &testreq.FakeReqFactory{}

	ui := callCreateSpace(t, []string{}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestCreateSpaceRequirements(t *testing.T) {
	resetSpaceDefaults()
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callCreateSpace(t, []string{"-o", "some-org", "my-space"}, reqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

}

func TestCreateSpace(t *testing.T) {
	resetSpaceDefaults()
	ui := callCreateSpace(t, []string{"my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

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

func TestCreateSpaceInOrg(t *testing.T) {
	resetSpaceDefaults()

	defaultOrgRepo.FindByNameOrganization = maker.NewOrg(maker.Overrides{
		"name": "other-org",
	})

	ui := callCreateSpace(t, []string{"-o", "other-org", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating space", "my-space", "other-org", "my-user"},
		{"OK"},
		{"Assigning", "my-user", "my-space", cf.SpaceRoleToUserInput[cf.SPACE_MANAGER]},
		{"Assigning", "my-user", "my-space", cf.SpaceRoleToUserInput[cf.SPACE_DEVELOPER]},
		{"TIP"},
	})

	assert.Equal(t, defaultSpaceRepo.CreateSpaceName, "my-space")
	assert.Equal(t, defaultSpaceRepo.CreateSpaceOrgGuid, defaultOrgRepo.FindByNameOrganization.Guid)
	assert.Equal(t, defaultUserRepo.SetSpaceRoleUserGuid, "my-user-guid")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleSpaceGuid, "my-space-guid")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleRole, cf.SPACE_DEVELOPER)
}

func TestCreateSpaceInOrgWhenTheOrgDoesNotExist(t *testing.T) {
	resetSpaceDefaults()

	defaultOrgRepo.FindByNameNotFound = true

	ui := callCreateSpace(t, []string{"-o", "cool-organization", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"cool-organization", "does not exist"},
	})

	assert.Equal(t, defaultSpaceRepo.CreateSpaceName, "")
}

func TestCreateSpaceInOrgWhenErrorFindingOrg(t *testing.T) {
	resetSpaceDefaults()

	defaultOrgRepo.FindByNameErr = true

	ui := callCreateSpace(t, []string{"-o", "cool-organization", "my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"error"},
	})

	assert.Equal(t, defaultSpaceRepo.CreateSpaceName, "")
}

func TestCreateSpaceWhenItAlreadyExists(t *testing.T) {
	resetSpaceDefaults()
	defaultSpaceRepo.CreateSpaceExists = true
	ui := callCreateSpace(t, []string{"my-space"}, defaultReqFactory, defaultSpaceRepo, defaultOrgRepo, defaultUserRepo)

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

func callCreateSpace(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository, orgRepo *testapi.FakeOrgRepository, userRepo *testapi.FakeUserRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
		UserGuid: "my-user-guid",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		SpaceFields:        configSpace,
		OrganizationFields: configOrg,
		AccessToken:        token,
	}

	spaceRoleSetter := user.NewSetSpaceRole(ui, config, spaceRepo, userRepo)
	cmd := NewCreateSpace(ui, config, spaceRoleSetter, spaceRepo, orgRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
