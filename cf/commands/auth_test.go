package commands_test

import (
	"github.com/cloudfoundry/cli/cf"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("auth command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		repo                *testapi.FakeAuthenticationRepository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(repo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("auth").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		repo = &testapi.FakeAuthenticationRepository{
			Config:       config,
			AccessToken:  "my-access-token",
			RefreshToken: "my-refresh-token",
		}

		deps = command_registry.NewDependency()
	})

	Describe("requirements", func() {
		It("fails with usage when given too few arguments", func() {
			testcmd.RunCliCommand("auth", []string{}, requirementsFactory, updateCommandDependency, false)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails if the user has not set an api endpoint", func() {
			Expect(testcmd.RunCliCommand("auth", []string{"username", "password"}, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})
	})

	Context("when an api endpoint is targeted", func() {
		BeforeEach(func() {
			requirementsFactory.ApiEndpointSuccess = true
			config.SetApiEndpoint("foo.example.org/authenticate")
		})

		It("authenticates successfully", func() {
			requirementsFactory.ApiEndpointSuccess = true
			testcmd.RunCliCommand("auth", []string{"foo@example.com", "password"}, requirementsFactory, updateCommandDependency, false)

			Expect(ui.FailedWithUsage).To(BeFalse())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"foo.example.org/authenticate"},
				[]string{"OK"},
			))

			Expect(repo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
				{
					"username": "foo@example.com",
					"password": "password",
				},
			}))
		})

		It("prompts users to upgrade if CLI version < min cli version requirement", func() {
			config.SetMinCliVersion("5.0.0")
			config.SetMinRecommendedCliVersion("5.5.0")
			cf.Version = "4.5.0"

			testcmd.RunCliCommand("auth", []string{"foo@example.com", "password"}, requirementsFactory, updateCommandDependency, false)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"To upgrade your CLI"},
				[]string{"5.0.0"},
			))
		})

		It("gets the UAA endpoint and saves it to the config file", func() {
			requirementsFactory.ApiEndpointSuccess = true
			testcmd.RunCliCommand("auth", []string{"foo@example.com", "password"}, requirementsFactory, updateCommandDependency, false)
			Expect(repo.GetLoginPromptsWasCalled).To(BeTrue())
		})

		Describe("when authentication fails", func() {
			BeforeEach(func() {
				repo.AuthError = true
				testcmd.RunCliCommand("auth", []string{"username", "password"}, requirementsFactory, updateCommandDependency, false)
			})

			It("does not prompt the user when provided username and password", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{config.ApiEndpoint()},
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
