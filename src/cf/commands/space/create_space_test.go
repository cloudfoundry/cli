package space_test

import (
	"cf"
	. "cf/commands/space"
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
	defaultSpace     cf.Space
	configSpace      cf.SpaceFields
	defaultSpaceRepo *testapi.FakeSpaceRepository
	defaultUserRepo  *testapi.FakeUserRepository
)

func resetSpaceDefaults() {
	defaultSpace = cf.Space{}
	defaultSpace.Name = "my-space"
	defaultSpace.Guid = "my-space-guid"

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

	fakeUI := callCreateSpace(t, []string{}, reqFactory, defaultSpaceRepo, defaultUserRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)
	assert.False(t, fakeUI.FailedWithUsage)
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
	fakeUI := callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)

	testassert.SliceContains(t, fakeUI.Outputs, testassert.Lines{
		{"Creating space", "my-space", "my-org", "my-user"},
		{"OK"},
		{"Binding", "my-user", "my-space", cf.SpaceRoleToUserInput[cf.SPACE_MANAGER]},
		{"Binding", "my-user", "my-space", cf.SpaceRoleToUserInput[cf.SPACE_DEVELOPER]},
		{"TIP"},
	})

	assert.Equal(t, defaultSpaceRepo.CreateSpaceName, "my-space")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleUserGuid, "my-user-guid")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleSpaceGuid, "my-space-guid")
	assert.Equal(t, defaultUserRepo.SetSpaceRoleRole, cf.SPACE_DEVELOPER)
}

func TestCreateSpaceWhenItAlreadyExists(t *testing.T) {
	resetSpaceDefaults()
	defaultSpaceRepo.CreateSpaceExists = true

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	fakeUI := callCreateSpace(t, []string{"my-space"}, reqFactory, defaultSpaceRepo, defaultUserRepo)

	assert.Equal(t, len(fakeUI.Outputs), 3)
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-space")
	assert.Contains(t, fakeUI.Outputs[2], "already exists")
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
	config := &configuration.Configuration{
		SpaceFields:        configSpace,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewCreateSpace(ui, config, spaceRepo, userRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
