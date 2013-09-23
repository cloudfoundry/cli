package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestRestartCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	starter := &testhelpers.FakeAppStarter{}
	stopper := &testhelpers.FakeAppStopper{}
	ui := callRestart([]string{}, reqFactory, starter, stopper)
	assert.True(t, ui.FailedWithUsage)

	ui = callRestart([]string{"my-app"}, reqFactory, starter, stopper)
	assert.False(t, ui.FailedWithUsage)
}

func TestRestartRequirements(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	starter := &testhelpers.FakeAppStarter{}
	stopper := &testhelpers.FakeAppStopper{}

	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	callRestart([]string{"my-app"}, reqFactory, starter, stopper)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
	callRestart([]string{"my-app"}, reqFactory, starter, stopper)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
	callRestart([]string{"my-app"}, reqFactory, starter, stopper)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestRestartApplication(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	stoppedApp := cf.Application{Name: "my-stopped-app", Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	starter := &testhelpers.FakeAppStarter{}
	stopper := &testhelpers.FakeAppStopper{StoppedApp: stoppedApp}
	callRestart([]string{"my-app"}, reqFactory, starter, stopper)

	assert.Equal(t, stopper.AppToStop, app)
	assert.Equal(t, starter.AppToStart, stoppedApp)
}

func callRestart(args []string, reqFactory *testhelpers.FakeReqFactory, starter ApplicationStarter, stopper ApplicationStopper) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("restart", args)

	cmd := NewRestart(ui, starter, stopper)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
