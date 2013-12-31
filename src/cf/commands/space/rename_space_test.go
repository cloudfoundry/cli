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

func TestRenameSpaceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{}

	ui := callRenameSpace(t, []string{}, reqFactory, spaceRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRenameSpace(t, []string{"foo"}, reqFactory, spaceRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestRenameSpaceRequirements(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callRenameSpace(t, []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callRenameSpace(t, []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	callRenameSpace(t, []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.SpaceName, "my-space")
}

func TestRenameSpaceRun(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}
	space := cf.Space{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: space}
	ui := callRenameSpace(t, []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Renaming space", "my-space", "my-new-space", "my-org", "my-user"},
		{"OK"},
	})

	assert.Equal(t, spaceRepo.RenameSpaceGuid, "my-space-guid")
	assert.Equal(t, spaceRepo.RenameNewName, "my-new-space")
}

func callRenameSpace(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space2 := cf.SpaceFields{}
	space2.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space2,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewRenameSpace(ui, config, spaceRepo, testconfig.FakeConfigRepository{})
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
