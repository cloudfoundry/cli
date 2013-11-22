package application_test

import (
	"cf"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRestartCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	starter := &testcmd.FakeAppStarter{}
	stopper := &testcmd.FakeAppStopper{}
	ui := callRestart([]string{}, reqFactory, starter, stopper)
	assert.True(t, ui.FailedWithUsage)

	ui = callRestart([]string{"my-app"}, reqFactory, starter, stopper)
	assert.False(t, ui.FailedWithUsage)
}

func TestRestartRequirements(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	starter := &testcmd.FakeAppStarter{}
	stopper := &testcmd.FakeAppStopper{}

	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	callRestart([]string{"my-app"}, reqFactory, starter, stopper)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
	callRestart([]string{"my-app"}, reqFactory, starter, stopper)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
	callRestart([]string{"my-app"}, reqFactory, starter, stopper)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestRestartApplication(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	starter := &testcmd.FakeAppStarter{}
	stopper := &testcmd.FakeAppStopper{}
	callRestart([]string{"my-app"}, reqFactory, starter, stopper)

	assert.Equal(t, stopper.AppToStop, app)
	assert.Equal(t, starter.AppToStart, app)
}

func callRestart(args []string, reqFactory *testreq.FakeReqFactory, starter ApplicationStarter, stopper ApplicationStopper) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("restart", args)

	cmd := NewRestart(ui, starter, stopper)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
