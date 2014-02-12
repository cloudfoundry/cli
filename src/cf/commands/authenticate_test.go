package commands_test

import (
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callAuthenticate([]string{"my-username"}, config, auth)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callAuthenticate([]string{"my-username", "my-password"}, config, auth)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestSuccessfullyAuthenticatingWithUsernameAndPasswordAsArguments", func() {
		testSuccessfulAuthenticate([]string{"user@example.com", "password"})
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

		Expect(len(ui.Outputs)).To(Equal(4))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{config.ApiEndpoint()},
			{"Authenticating..."},
			{"FAILED"},
			{"Error authenticating"},
		})
	})
})

func testSuccessfulAuthenticate(args []string) (ui *testterm.FakeUI) {
	config := testconfig.NewRepository()
	config.SetApiEndpoint("foo.example.org/authenticate")

	auth := &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		Config:       config,
	}

	ui = callAuthenticate(args, config, auth)

	testassert.SliceContains(ui.Outputs, testassert.Lines{
		{"foo.example.org/authenticate"},
		{"OK"},
	})

	Expect(config.AccessToken()).To(Equal("my_access_token"))
	Expect(config.RefreshToken()).To(Equal("my_refresh_token"))
	Expect(auth.Email).To(Equal("user@example.com"))
	Expect(auth.Password).To(Equal("password"))

	return
}

func callAuthenticate(args []string, config configuration.Reader, auth api.AuthenticationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("auth", args)
	cmd := NewAuthenticate(ui, config, auth)
	testcmd.RunCommand(cmd, ctxt, &testreq.FakeReqFactory{})
	return
}
