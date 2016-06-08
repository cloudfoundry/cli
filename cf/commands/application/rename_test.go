package application_test

import (
	"github.com/cloudfoundry/cli/cf/api/applications/applicationsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rename command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		appRepo             *applicationsfakes.FakeRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("rename").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		appRepo = new(applicationsfakes.FakeRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("rename", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when not invoked with an old name and a new name", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand("foo")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-app", "my-new-app")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(runCommand("my-app", "my-new-app")).To(BeFalse())
		})
	})

	It("renames an application", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})

		app := models.Application{}
		app.Name = "my-app"
		app.GUID = "my-app-guid"
		applicationReq := new(requirementsfakes.FakeApplicationRequirement)
		applicationReq.GetApplicationReturns(app)
		requirementsFactory.NewApplicationRequirementReturns(applicationReq)

		runCommand("my-app", "my-new-app")

		appGUID, params := appRepo.UpdateArgsForCall(0)
		Expect(appGUID).To(Equal(app.GUID))
		Expect(*params.Name).To(Equal("my-new-app"))
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Renaming app", "my-app", "my-new-app", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
	})
})
