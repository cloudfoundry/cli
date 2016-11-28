package application_test

import (
	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("restage command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *applicationsfakes.FakeRepository
		configRepo          coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		stagingWatcher      *fakeStagingWatcher
		OriginalCommand     commandregistry.Command
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.Config = configRepo

		//inject fake 'command dependency' into registry
		commandregistry.Register(stagingWatcher)

		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("restage").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		app = models.Application{}
		app.Name = "my-app"
		app.PackageState = "STAGED"
		appRepo = new(applicationsfakes.FakeRepository)
		appRepo.ReadReturns(app, nil)

		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})

		//save original command and restore later
		OriginalCommand = commandregistry.Commands.FindCommand("start")

		stagingWatcher = &fakeStagingWatcher{}
	})

	AfterEach(func() {
		commandregistry.Register(OriginalCommand)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("restage", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("Requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails with usage when no arguments are given", func() {
			passed := runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
			Expect(passed).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(runCommand("my-app")).To(BeFalse())
		})
	})

	It("fails with usage when the app cannot be found", func() {
		appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("app", "hocus-pocus"))
		runCommand("hocus-pocus")

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"not found"},
		))
	})

	Context("when the app is found", func() {
		BeforeEach(func() {
			app = models.Application{}
			app.Name = "my-app"
			app.GUID = "the-app-guid"

			appRepo.ReadReturns(app, nil)
		})

		It("sends a restage request", func() {
			runCommand("my-app")
			Expect(appRepo.CreateRestageRequestArgsForCall(0)).To(Equal("the-app-guid"))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Restaging app", "my-app", "my-org", "my-space", "my-user"},
			))
		})

		It("resets app's PackageState", func() {
			runCommand("my-app")
			Expect(stagingWatcher.watched.PackageState).ToNot(Equal("STAGED"))
		})

		It("watches the staging output", func() {
			runCommand("my-app")
			Expect(stagingWatcher.watched).To(Equal(app))
			Expect(stagingWatcher.orgName).To(Equal(configRepo.OrganizationFields().Name))
			Expect(stagingWatcher.spaceName).To(Equal(configRepo.SpaceFields().Name))
		})
	})
})

type fakeStagingWatcher struct {
	watched   models.Application
	orgName   string
	spaceName string
}

func (f *fakeStagingWatcher) WatchStaging(app models.Application, orgName, spaceName string, start func(models.Application) (models.Application, error)) (updatedApp models.Application, err error) {
	f.watched = app
	f.orgName = orgName
	f.spaceName = spaceName
	return start(app)
}
func (cmd *fakeStagingWatcher) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{Name: "start"}
}

func (cmd *fakeStagingWatcher) SetDependency(_ commandregistry.Dependency, _ bool) commandregistry.Command {
	return cmd
}

func (cmd *fakeStagingWatcher) Requirements(_ requirements.Factory, _ flags.FlagContext) ([]requirements.Requirement, error) {
	return []requirements.Requirement{}, nil
}

func (cmd *fakeStagingWatcher) Execute(_ flags.FlagContext) error {
	return nil
}
