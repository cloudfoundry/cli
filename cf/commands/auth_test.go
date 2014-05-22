package commands_test

import (
	. "github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
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
		cmd                 Authenticate
		config              configuration.ReadWriter
		repo                *testapi.FakeAuthenticationRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		repo = &testapi.FakeAuthenticationRepository{
			Config:       config,
			AccessToken:  "my-access-token",
			RefreshToken: "my-refresh-token",
		}
		cmd = NewAuthenticate(ui, config, repo)
	})

	Describe("requirements", func() {
		It("fails with usage when given too few arguments", func() {
			testcmd.RunCommand(cmd, []string{}, requirementsFactory)

			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails if the user has not set an api endpoint", func() {
			testcmd.RunCommand(cmd, []string{"username", "password"}, requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when an api endpoint is targeted", func() {
		BeforeEach(func() {
			requirementsFactory.ApiEndpointSuccess = true
			config.SetApiEndpoint("foo.example.org/authenticate")
		})

		It("authenticates successfully", func() {
			requirementsFactory.ApiEndpointSuccess = true
			testcmd.RunCommand(cmd, []string{"foo@example.com", "password"}, requirementsFactory)

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

		It("gets the UAA endpoint and saves it to the config file", func() {
			requirementsFactory.ApiEndpointSuccess = true
			testcmd.RunCommand(cmd, []string{"foo@example.com", "password"}, requirementsFactory)
			Expect(repo.GetLoginPromptsWasCalled).To(BeTrue())
		})

		Describe("when authentication fails", func() {
			BeforeEach(func() {
				repo.AuthError = true
				testcmd.RunCommand(cmd, []string{"username", "password"}, requirementsFactory)
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
