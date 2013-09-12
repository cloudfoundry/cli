package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestRenameSpaceFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	fakeUI := callRenameSpace([]string{}, reqFactory, spaceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRenameSpace([]string{"foo"}, reqFactory, spaceRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameSpaceRequirements(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callRenameSpace([]string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.SpaceName, "my-space")
}

func TestRenameSpaceRun(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	space := cf.Space{Name: "my-space", Guid: "my-space-guid"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Space: space}
	ui := callRenameSpace([]string{"my-space", "my-new-space"}, reqFactory, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "Renaming space")
	assert.Equal(t, spaceRepo.RenameSpace, space)
	assert.Equal(t, spaceRepo.RenameNewName, "my-new-space")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRenameSpace(args []string, reqFactory *testhelpers.FakeReqFactory, spaceRepo *testhelpers.FakeSpaceRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("create-space", args)
	cmd := NewRenameSpace(ui, spaceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
