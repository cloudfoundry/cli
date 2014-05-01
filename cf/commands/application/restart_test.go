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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func callRestart(args []string, requirementsFactory *testreq.FakeReqFactory, starter ApplicationStarter, stopper ApplicationStopper) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("restart", args)

	cmd := NewRestart(ui, starter, stopper)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestRestartCommandFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		starter := &testcmd.FakeAppStarter{}
		stopper := &testcmd.FakeAppStopper{}
		ui := callRestart([]string{}, requirementsFactory, starter, stopper)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callRestart([]string{"my-app"}, requirementsFactory, starter, stopper)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestRestartRequirements", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		starter := &testcmd.FakeAppStarter{}
		stopper := &testcmd.FakeAppStopper{}

		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		callRestart([]string{"my-app"}, requirementsFactory, starter, stopper)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
		callRestart([]string{"my-app"}, requirementsFactory, starter, stopper)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
		callRestart([]string{"my-app"}, requirementsFactory, starter, stopper)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestRestartApplication", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		requirementsFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		starter := &testcmd.FakeAppStarter{}
		stopper := &testcmd.FakeAppStopper{}
		callRestart([]string{"my-app"}, requirementsFactory, starter, stopper)

		Expect(stopper.AppToStop).To(Equal(app))
		Expect(starter.AppToStart).To(Equal(app))
	})
})
