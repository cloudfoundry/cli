package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"

	. "testhelpers/matchers"
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
			context := testcmd.NewContext("auth", []string{})
			testcmd.RunCommand(cmd, context, requirementsFactory)

			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails if the user has not set an api endpoint", func() {
			context := testcmd.NewContext("auth", []string{"username", "password"})
			testcmd.RunCommand(cmd, context, requirementsFactory)

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
			context := testcmd.NewContext("auth", []string{"foo@example.com", "password"})
			testcmd.RunCommand(cmd, context, requirementsFactory)

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
			testcmd.RunCommand(cmd, testcmd.NewContext("auth", []string{"foo@example.com", "password"}), requirementsFactory)
			Expect(repo.GetLoginPromptsWasCalled).To(BeTrue())
		})

		Describe("when authentication fails", func() {
			BeforeEach(func() {
				repo.AuthError = true
				testcmd.RunCommand(cmd, testcmd.NewContext("auth", []string{"username", "password"}), requirementsFactory)
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
