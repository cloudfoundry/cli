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

var _ = Describe("Testing with ginkgo", func() {
	It("TestAuthenticateFailsWithUsage", func() {
		config := testconfig.NewRepository()

		auth := &testapi.FakeAuthenticationRepository{
			AccessToken:  "my_access_token",
			RefreshToken: "my_refresh_token",
			Config:       config,
		}

		ui := callAuthenticate([]string{}, config, auth)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callAuthenticate([]string{"my-username"}, config, auth)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callAuthenticate([]string{"my-username", "my-password"}, config, auth)
		assert.False(mr.T(), ui.FailedWithUsage)
	})

	It("TestSuccessfullyAuthenticatingWithUsernameAndPasswordAsArguments", func() {
		testSuccessfulAuthenticate(mr.T(), []string{"user@example.com", "password"})
	})

	It("TestUnsuccessfullyAuthenticatingWithoutInteractivity", func() {
		config := testconfig.NewRepository()

		ui := callAuthenticate(
			[]string{
				"foo@example.com",
				"bar",
			},
			config,
			&testapi.FakeAuthenticationRepository{AuthError: true, Config: config},
		)

		assert.Equal(mr.T(), len(ui.Outputs), 4)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{config.ApiEndpoint()},
			{"Authenticating..."},
			{"FAILED"},
			{"Error authenticating"},
		})
	})
})

func testSuccessfulAuthenticate(t mr.TestingT, args []string) (ui *testterm.FakeUI) {
	config := testconfig.NewRepository()
	config.SetApiEndpoint("foo.example.org/authenticate")

	auth := &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		Config:       config,
	}

	ui = callAuthenticate(args, config, auth)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"foo.example.org/authenticate"},
		{"OK"},
	})

	assert.Equal(t, config.AccessToken(), "my_access_token")
	assert.Equal(t, config.RefreshToken(), "my_refresh_token")
	assert.Equal(t, auth.Email, "user@example.com")
	assert.Equal(t, auth.Password, "password")

	return
}

func callAuthenticate(args []string, config configuration.Reader, auth api.AuthenticationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("auth", args)
	cmd := NewAuthenticate(ui, config, auth)
	testcmd.RunCommand(cmd, ctxt, &testreq.FakeReqFactory{})
	return
}
