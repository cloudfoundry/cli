package commands_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/authentication/authenticationfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("auth command", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		authRepo            *authenticationfakes.FakeRepository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
		fakeLogger          *tracefakes.FakePrinter
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("auth").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		authRepo = new(authenticationfakes.FakeRepository)
		authRepo.AuthenticateStub = func(credentials map[string]string) error {
			config.SetAccessToken("my-access-token")
			config.SetRefreshToken("my-refresh-token")
			return nil
		}

		fakeLogger = new(tracefakes.FakePrinter)
		deps = commandregistry.NewDependency(os.Stdout, fakeLogger, "")
	})

	Describe("requirements", func() {
		It("fails with usage when given too few arguments", func() {
			testcmd.RunCLICommand("auth", []string{}, requirementsFactory, updateCommandDependency, false, ui)

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails if the user has not set an api endpoint", func() {
			requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Failing{Message: "no api set"})
			Expect(testcmd.RunCLICommand("auth", []string{"username", "password"}, requirementsFactory, updateCommandDependency, false, ui)).To(BeFalse())
		})
	})

	Context("when an api endpoint is targeted", func() {
		BeforeEach(func() {
			requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Passing{})
			config.SetAPIEndpoint("foo.example.org/authenticate")
		})

		It("authenticates successfully", func() {
			requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Passing{})
			testcmd.RunCLICommand("auth", []string{"foo@example.com", "password"}, requirementsFactory, updateCommandDependency, false, ui)

			Expect(ui.FailedWithUsage).To(BeFalse())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"foo.example.org/authenticate"},
				[]string{"OK"},
			))

			Expect(authRepo.AuthenticateArgsForCall(0)).To(Equal(map[string]string{
				"username": "foo@example.com",
				"password": "password",
			}))
		})

		It("displays an update notification", func() {
			testcmd.RunCLICommand("auth", []string{"foo@example.com", "password"}, requirementsFactory, updateCommandDependency, false, ui)
			Expect(ui.NotifyUpdateIfNeededCallCount).To(Equal(1))
		})

		It("gets the UAA endpoint and saves it to the config file", func() {
			requirementsFactory.NewAPIEndpointRequirementReturns(requirements.Passing{})
			testcmd.RunCLICommand("auth", []string{"foo@example.com", "password"}, requirementsFactory, updateCommandDependency, false, ui)
			Expect(authRepo.GetLoginPromptsAndSaveUAAServerURLCallCount()).To(Equal(1))
		})

		Describe("when authentication fails", func() {
			BeforeEach(func() {
				authRepo.AuthenticateReturns(errors.New("Error authenticating."))
				testcmd.RunCLICommand("auth", []string{"username", "password"}, requirementsFactory, updateCommandDependency, false, ui)
			})

			It("does not prompt the user when provided username and password", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{config.APIEndpoint()},
					[]string{"Authenticating..."},
					[]string{"FAILED"},
					[]string{"Error authenticating"},
				))
			})

			It("clears the user's session", func() {
				Expect(config.AccessToken()).To(BeEmpty())
				Expect(config.RefreshToken()).To(BeEmpty())
				Expect(config.SpaceFields()).To(Equal(models.SpaceFields{}))
				Expect(config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
			})
		})
	})
})
