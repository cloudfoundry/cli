package application_test

import (
	testappfiles "github.com/cloudfoundry/cli/cf/api/app_files/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
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
		configRepo          core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
		appFilesRepo        *testappfiles.FakeAppFilesRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		appFilesRepo = &testappfiles.FakeAppFilesRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewFiles(ui, configRepo, appFilesRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedSpaceSuccess = true
			runCommand("my-app", "/foo")
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			Expect(runCommand("my-app", "/foo")).To(BeFalse())
		})

		It("fails with usage when not provided an app name", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			passed := runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(passed).To(BeFalse())
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
			appFilesRepo.ListFilesReturns("file 1\nfile 2", nil)
		})

		It("it lists files in a directory", func() {
			runCommand("my-app", "/foo")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting files for app", "my-found-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"file 1"},
				[]string{"file 2"},
			))

			guid, _, path := appFilesRepo.ListFilesArgsForCall(0)
			Expect(guid).To(Equal("my-app-guid"))
			Expect(path).To(Equal("/foo"))
		})

		It("does not interpolate or interpret special format characters as though it should be a format string", func() {
			appFilesRepo.ListFilesReturns("%s %d %i", nil)
			runCommand("my-app", "/foo")

			Expect(ui.Outputs).To(ContainSubstrings([]string{"%s %d %i"}))
		})

		Context("checking for bad flags", func() {
			It("fails when non-positive value is given for instance", func() {
				runCommand("-i", "-1", "my-app")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Invalid instance"},
					[]string{"Instance must be a positive integer"},
				))
			})

			It("fails when instance is larger than instance count", func() {
				runCommand("-i", "5", "my-app")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Invalid instance"},
					[]string{"Instance must be less than"},
				))
			})

		})

		Context("when there is no file to be listed", func() {
			BeforeEach(func() {
				appFilesRepo.ListFilesReturns("", nil)
			})

			It("informs user that the directory is empty", func() {
				runCommand("my-app", "/foo")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting files for app", "my-found-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{""},
					[]string{"No files found"},
				))
			})

		})

	})
})
