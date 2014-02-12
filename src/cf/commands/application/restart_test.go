package application_test

import (
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callRestart(args []string, reqFactory *testreq.FakeReqFactory, starter ApplicationStarter, stopper ApplicationStopper) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("restart", args)

	cmd := NewRestart(ui, starter, stopper)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestRestartCommandFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}
		starter := &testcmd.FakeAppStarter{}
		stopper := &testcmd.FakeAppStopper{}
		ui := callRestart([]string{}, reqFactory, starter, stopper)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callRestart([]string{"my-app"}, reqFactory, starter, stopper)
		assert.False(mr.T(), ui.FailedWithUsage)
	})
	It("TestRestartRequirements", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		starter := &testcmd.FakeAppStarter{}
		stopper := &testcmd.FakeAppStopper{}

		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		callRestart([]string{"my-app"}, reqFactory, starter, stopper)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
		callRestart([]string{"my-app"}, reqFactory, starter, stopper)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
		callRestart([]string{"my-app"}, reqFactory, starter, stopper)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)
	})
	It("TestRestartApplication", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		starter := &testcmd.FakeAppStarter{}
		stopper := &testcmd.FakeAppStopper{}
		callRestart([]string{"my-app"}, reqFactory, starter, stopper)

		assert.Equal(mr.T(), stopper.AppToStop, app)
		assert.Equal(mr.T(), starter.AppToStart, app)
	})
})
