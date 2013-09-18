package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestRenameAppFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	appRepo := &testhelpers.FakeApplicationRepository{}

	fakeUI := callRename([]string{}, reqFactory, appRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRename([]string{"foo"}, reqFactory, appRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameRequirements(t *testing.T) {
	appRepo := &testhelpers.FakeApplicationRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callRename([]string{"my-app", "my-new-app"}, reqFactory, appRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestRenameRun(t *testing.T) {
	appRepo := &testhelpers.FakeApplicationRepository{}

	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Application: app}
	ui := callRename([]string{"my-app", "my-new-app"}, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Renaming ")
	assert.Equal(t, appRepo.RenameApp, app)
	assert.Equal(t, appRepo.RenameNewName, "my-new-app")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRename(args []string, reqFactory *testhelpers.FakeReqFactory, appRepo *testhelpers.FakeApplicationRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("rename", args)
	cmd := NewRename(ui, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
