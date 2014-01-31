package commands_test

import (
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func testSuccessfulAuthenticate(t mr.TestingT, args []string) (ui *testterm.FakeUI) {
	configRepo := testconfig.FakeConfigRepository{}
	configRepo.Delete()
	config, _ := configRepo.Get()

	auth := &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		ConfigRepo:   configRepo,
	}
	ui = callAuthenticate(
		args,
		configRepo,
		auth,
	)

	savedConfig := testconfig.SavedConfiguration

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{config.Target},
		{"OK"},
	})

	assert.Equal(t, savedConfig.AccessToken, "my_access_token")
	assert.Equal(t, savedConfig.RefreshToken, "my_refresh_token")
	assert.Equal(t, auth.Email, "user@example.com")
	assert.Equal(t, auth.Password, "password")

	return
}

func callAuthenticate(args []string, configRepo configuration.ConfigurationRepository, auth api.AuthenticationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("auth", args)
	cmd := NewAuthenticate(ui, configRepo, auth)
	testcmd.RunCommand(cmd, ctxt, &testreq.FakeReqFactory{})
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestAuthenticateFailsWithUsage", func() {
			configRepo := testconfig.FakeConfigRepository{}
			configRepo.Delete()

			auth := &testapi.FakeAuthenticationRepository{
				AccessToken:  "my_access_token",
				RefreshToken: "my_refresh_token",
				ConfigRepo:   configRepo,
			}

			ui := callAuthenticate([]string{}, configRepo, auth)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callAuthenticate([]string{"my-username"}, configRepo, auth)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callAuthenticate([]string{"my-username", "my-password"}, configRepo, auth)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestSuccessfullyAuthenticatingWithUsernameAndPasswordAsArguments", func() {

			testSuccessfulAuthenticate(mr.T(), []string{"user@example.com", "password"})
		})
		It("TestUnsuccessfullyAuthenticatingWithoutInteractivity", func() {

			configRepo := testconfig.FakeConfigRepository{}
			configRepo.Delete()
			config, _ := configRepo.Get()

			ui := callAuthenticate(
				[]string{
					"foo@example.com",
					"bar",
				},
				configRepo,
				&testapi.FakeAuthenticationRepository{AuthError: true, ConfigRepo: configRepo},
			)

			assert.Equal(mr.T(), len(ui.Outputs), 4)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{config.Target},
				{"Authenticating..."},
				{"FAILED"},
				{"Error authenticating"},
			})
		})
	})
}
