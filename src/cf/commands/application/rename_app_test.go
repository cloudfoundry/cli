package application_test

import (
	"cf"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRenameAppFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	appRepo := &testapi.FakeApplicationRepository{}

	fakeUI := callRename([]string{}, reqFactory, appRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRename([]string{"foo"}, reqFactory, appRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameRequirements(t *testing.T) {
	appRepo := &testapi.FakeApplicationRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callRename([]string{"my-app", "my-new-app"}, reqFactory, appRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestRenameRun(t *testing.T) {
	appRepo := &testapi.FakeApplicationRepository{}

	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	ui := callRename([]string{"my-app", "my-new-app"}, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Renaming ")
	assert.Equal(t, appRepo.RenameApp, app)
	assert.Equal(t, appRepo.RenameNewName, "my-new-app")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRename(args []string, reqFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename", args)
	cmd := NewRenameApp(ui, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
