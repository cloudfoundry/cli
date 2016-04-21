package application_test

import (
	"github.com/cloudfoundry/cli/cf/api/applications/applicationsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rename command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          coreconfig.Repository
		appRepo             *applicationsfakes.FakeApplicationRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("rename").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		appRepo = new(applicationsfakes.FakeApplicationRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("rename", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when not invoked with an old name and a new name", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("foo")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			Expect(runCommand("my-app", "my-new-app")).To(BeFalse())
		})
		It("fails if a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("my-app", "my-new-app")).To(BeFalse())
		})
	})

	It("renames an application", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.GUID = "my-app-guid"
		requirementsFactory.LoginSuccess = true
		requirementsFactory.TargetedSpaceSuccess = true
		requirementsFactory.Application = app
		runCommand("my-app", "my-new-app")

		appGUID, params := appRepo.UpdateArgsForCall(0)
		Expect(appGUID).To(Equal(app.GUID))
		Expect(*params.Name).To(Equal("my-new-app"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Renaming app", "my-app", "my-new-app", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
	})
})
