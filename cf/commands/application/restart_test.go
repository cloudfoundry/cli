package application_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("restart command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		starter             *testcmd.FakeApplicationStarter
		stopper             *testcmd.FakeApplicationStopper
		config              core_config.ReadWriter
		app                 models.Application
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		starter = &testcmd.FakeApplicationStarter{}
		stopper = &testcmd.FakeApplicationStopper{}
		config = testconfig.NewRepositoryWithDefaults()

		app = models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewRestart(ui, config, starter, stopper), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.Application = app
			requirementsFactory.TargetedSpaceSuccess = true

			Expect(runCommand()).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true

			Expect(runCommand()).To(BeFalse())
		})
	})

	Context("when logged in, targeting a space, and an app name is provided", func() {
		BeforeEach(func() {
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			stopper.ApplicationStopReturns(app, nil)
		})

		It("restarts the given app", func() {
			runCommand("my-app")

			application, orgName, spaceName := stopper.ApplicationStopArgsForCall(0)
			Expect(application).To(Equal(app))
			Expect(orgName).To(Equal(config.OrganizationFields().Name))
			Expect(spaceName).To(Equal(config.SpaceFields().Name))

			application, orgName, spaceName = starter.ApplicationStartArgsForCall(0)
			Expect(application).To(Equal(app))
			Expect(orgName).To(Equal(config.OrganizationFields().Name))
			Expect(spaceName).To(Equal(config.SpaceFields().Name))

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		})
	})
})
