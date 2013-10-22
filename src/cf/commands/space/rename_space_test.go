package space_test

import (
	"cf"
	. "cf/commands/space"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRenameSpaceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	spaceRepo := &testapi.FakeSpaceRepository{}

	fakeUI := callRenameSpace(t, []string{}, reqFactory, spaceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRenameSpace(t, []string{"foo"}, reqFactory, spaceRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameSpaceRequirements(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callRenameSpace(t, []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.SpaceName, "my-space")
}

func TestRenameSpaceRun(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Space: space}
	ui := callRenameSpace(t, []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "Renaming space")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-new-space")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Equal(t, spaceRepo.RenameSpace, space)
	assert.Equal(t, spaceRepo.RenameNewName, "my-new-space")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRenameSpace(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewRenameSpace(ui, config, spaceRepo, testconfig.FakeConfigRepository{})
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
