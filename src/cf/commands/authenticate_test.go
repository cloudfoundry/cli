package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("auth command", func() {
	var (
		ui         *testterm.FakeUI
		cmd        Authenticate
		config     configuration.ReadWriter
		repo       *testapi.FakeAuthenticationRepository
		reqFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		reqFactory = &testreq.FakeReqFactory{}
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
			testcmd.RunCommand(cmd, context, reqFactory)

			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails if the user has not set an api endpoint", func() {
			context := testcmd.NewContext("auth", []string{"username", "password"})
			testcmd.RunCommand(cmd, context, reqFactory)

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when an api endpoint is targeted", func() {
		BeforeEach(func() {
			reqFactory.ApiEndpointSuccess = true
			config.SetApiEndpoint("foo.example.org/authenticate")
		})

		It("authenticates successfully", func() {
			reqFactory.ApiEndpointSuccess = true
			context := testcmd.NewContext("auth", []string{"foo@example.com", "password"})
			testcmd.RunCommand(cmd, context, reqFactory)

			Expect(ui.FailedWithUsage).To(BeFalse())
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"foo.example.org/authenticate"},
				{"OK"},
			})

			Expect(repo.AuthenticateArgs.Credentials).To(Equal([]map[string]string{
				{
					"username": "foo@example.com",
					"password": "password",
				},
			}))
		})

		Describe("when authentication fails", func() {
			BeforeEach(func() {
				repo.AuthError = true
				testcmd.RunCommand(cmd, testcmd.NewContext("auth", []string{"username", "password"}), reqFactory)
			})

			It("does not prompt the user when provided username and password", func() {
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{config.ApiEndpoint()},
					{"Authenticating..."},
					{"FAILED"},
					{"Error authenticating"},
				})
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
