package application_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("files command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
		appFilesRepo        *testapi.FakeAppFilesRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		appFilesRepo = &testapi.FakeAppFilesRepo{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewFiles(ui, configRepo, appFilesRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedSpaceSuccess = true
			runCommand("my-app", "/foo")
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("my-app", "/foo")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when not provided an app name", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			runCommand()

			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in, a space is targeted and a valid app name is provided", func() {
		BeforeEach(func() {
			app := models.Application{}
			app.Name = "my-found-app"
			app.Guid = "my-app-guid"

			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			appFilesRepo.FileList = "file 1\nfile 2"
		})

		It("it lists files in a directory", func() {
			runCommand("my-app", "/foo")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting files for app", "my-found-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"file 1"},
				[]string{"file 2"},
			))

			Expect(appFilesRepo.AppGuid).To(Equal("my-app-guid"))
			Expect(appFilesRepo.Path).To(Equal("/foo"))
		})

		It("does not interpolate or interpret special format characters as though it should be a format string", func() {
			appFilesRepo.FileList = "%s %d %i"
			runCommand("my-app", "/foo")

			Expect(ui.Outputs).To(ContainSubstrings([]string{"%s %d %i"}))
		})
	})
})
