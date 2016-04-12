package application_test

import (
	"github.com/cloudfoundry/cli/cf/api/applications/applicationsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("restage command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *applicationsfakes.FakeApplicationRepository
		configRepo          coreconfig.Repository
		requirementsFactory *testreq.FakeReqFactory
		stagingWatcher      *fakeStagingWatcher
		OriginalCommand     commandregistry.Command
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
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
		appRepo = new(applicationsfakes.FakeApplicationRepository)
		appRepo.ReadReturns(app, nil)

		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

		//save original command and restore later
		OriginalCommand = commandregistry.Commands.FindCommand("start")

		stagingWatcher = &fakeStagingWatcher{}
	})

	AfterEach(func() {
		commandregistry.Register(OriginalCommand)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("restage", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("Requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails with usage when no arguments are given", func() {
			passed := runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
			Expect(passed).To(BeFalse())
		})
		It("fails if a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("my-app")).To(BeFalse())
		})
	})

	It("fails with usage when the app cannot be found", func() {
		appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("app", "hocus-pocus"))
		runCommand("hocus-pocus")

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"not found"},
		))
	})

	Context("when the app is found", func() {
		BeforeEach(func() {
			app = models.Application{}
			app.Name = "my-app"
			app.Guid = "the-app-guid"

			appRepo.ReadReturns(app, nil)
		})

		It("sends a restage request", func() {
			runCommand("my-app")
			Expect(appRepo.CreateRestageRequestArgsForCall(0)).To(Equal("the-app-guid"))
			Expect(ui.Outputs).To(ContainSubstrings(
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

func (f *fakeStagingWatcher) ApplicationWatchStaging(app models.Application, orgName, spaceName string, start func(models.Application) (models.Application, error)) (updatedApp models.Application, err error) {
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

func (cmd *fakeStagingWatcher) Requirements(_ requirements.Factory, _ flags.FlagContext) []requirements.Requirement {
	return []requirements.Requirement{}
}

func (cmd *fakeStagingWatcher) Execute(_ flags.FlagContext) {}
