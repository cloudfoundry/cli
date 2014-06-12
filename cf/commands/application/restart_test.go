package application_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("restart command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		starter             *testcmd.FakeAppStarter
		stopper             *testcmd.FakeAppStopper
		app                 models.Application
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		starter = &testcmd.FakeAppStarter{}
		stopper = &testcmd.FakeAppStopper{}

		app = models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewRestart(ui, starter, stopper), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when an app name is not given", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.Application = app
			requirementsFactory.TargetedSpaceSuccess = true
			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in, targeting a space, and an app name is provided", func() {
		BeforeEach(func() {
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
		})

		It("restarts the given app", func() {
			runCommand("my-app")

			Expect(stopper.AppToStop).To(Equal(app))
			Expect(starter.AppToStart).To(Equal(app))
			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		})
	})
})
