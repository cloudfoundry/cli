package application_test

import (
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callRestart([]string{"my-app"}, reqFactory, starter, stopper)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestRestartRequirements", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		starter := &testcmd.FakeAppStarter{}
		stopper := &testcmd.FakeAppStopper{}

		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		callRestart([]string{"my-app"}, reqFactory, starter, stopper)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
		callRestart([]string{"my-app"}, reqFactory, starter, stopper)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
		callRestart([]string{"my-app"}, reqFactory, starter, stopper)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestRestartApplication", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		starter := &testcmd.FakeAppStarter{}
		stopper := &testcmd.FakeAppStopper{}
		callRestart([]string{"my-app"}, reqFactory, starter, stopper)

		Expect(stopper.AppToStop).To(Equal(app))
		Expect(starter.AppToStart).To(Equal(app))
	})
})
