/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

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
