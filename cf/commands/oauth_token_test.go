package commands_test

import (
	"errors"

	authenticationfakes "github.com/cloudfoundry/cli/cf/api/authentication/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/trace/fakes"
	"github.com/cloudfoundry/cli/plugin/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OauthToken", func() {
	var (
		ui                  *testterm.FakeUI
		authRepo            *authenticationfakes.FakeAuthenticationRepository
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("oauth-token").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		fakeLogger := new(fakes.FakePrinter)
		authRepo = &authenticationfakes.FakeAuthenticationRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		deps = command_registry.NewDependency(fakeLogger)
	})

	runCommand := func() bool {
		return testcmd.RunCliCommand("oauth-token", []string{}, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails if oauth refresh fails", func() {
			authRepo.RefreshAuthTokenReturns("", errors.New("Could not refresh"))
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Could not refresh"},
			))
		})

		It("returns to the user the oauth token after a refresh", func() {
			authRepo.RefreshAuthTokenReturns("1234567890", nil)
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"1234567890"},
			))
		})

		Context("when invoked by a plugin", func() {
			var (
				pluginModel plugin_models.GetOauthToken_Model
			)

			BeforeEach(func() {
				pluginModel = plugin_models.GetOauthToken_Model{}
				deps.PluginModels.OauthToken = &pluginModel
			})

			It("populates the plugin model upon execution", func() {
				authRepo.RefreshAuthTokenReturns("911999111", nil)
				testcmd.RunCliCommand("oauth-token", []string{}, requirementsFactory, updateCommandDependency, true)
				Expect(pluginModel.Token).To(Equal("911999111"))
			})
		})
	})
})
