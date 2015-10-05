package commands_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/trace"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ssh-code command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		authRepo            *testapi.FakeAuthenticationRepository
		endpointRepo        *testapi.FakeEndpointRepo
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)
		deps.RepoLocator = deps.RepoLocator.SetEndpointRepository(endpointRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("ssh-code").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		authRepo = &testapi.FakeAuthenticationRepository{}
		endpointRepo = &testapi.FakeEndpointRepo{}

		deps = command_registry.NewDependency()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("ssh-code", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when invoked with any args", func() {
			runCommand("whoops")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "No argument required"},
			))
		})

		It("fails if the user has not set an api endpoint", func() {
			requirementsFactory.ApiEndpointSuccess = false

			Ω(runCommand()).To(BeFalse())
		})
	})

	Describe("ssh-code", func() {
		BeforeEach(func() {
			requirementsFactory.ApiEndpointSuccess = true
		})

		Context("calling endpoint repository to update 'app_ssh_oauth_client'", func() {
			It("passes the repo the targeted API endpoint", func() {
				configRepo.SetApiEndpoint("test.endpoint.com")

				runCommand()
				Ω(endpointRepo.CallCount).To(Equal(1))
				Ω(endpointRepo.UpdateEndpointReceived).To(Equal(configRepo.ApiEndpoint()))
			})

			It("reports any error to user", func() {
				configRepo.SetApiEndpoint("test.endpoint.com")
				endpointRepo.UpdateEndpointError = errors.New("endpoint error")

				runCommand()
				Ω(endpointRepo.CallCount).To(Equal(1))
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"Error getting info", "endpoint error"},
				))
			})
		})

		Context("refresh oauth-token to make sure it is not stale", func() {
			It("refreshes the oauth token to make sure it is not stale", func() {
				runCommand()
				Ω(authRepo.RefreshTokenCalled).To(BeTrue())
			})

			Context("when refreshing fails", func() {
				It("refreshes the oauth token to make sure it is not stale", func() {
					authRepo.RefreshTokenError = errors.New("no token for you!")

					runCommand()
					Ω(authRepo.RefreshTokenCalled).To(BeTrue())
					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"Error refreshing oauth token", "no token for you"},
					))
				})
			})
		})

		Context("setting up http client to request one time code", func() {
			var fakeUAA *ghttp.Server

			BeforeEach(func() {
				authRepo.RefreshToken = "bearer client-bearer-token"
				configRepo.SetSSLDisabled(true)
				configRepo.SetSSHOAuthClient("ssh-oauth-client-id")

				fakeUAA = ghttp.NewTLSServer()
				configRepo.SetUaaEndpoint(fakeUAA.URL())

				fakeUAA.RouteToHandler("GET", "/oauth/authorize", ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/oauth/authorize"),
					ghttp.VerifyFormKV("response_type", "code"),
					ghttp.VerifyFormKV("client_id", "ssh-oauth-client-id"),
					ghttp.VerifyFormKV("grant_type", "authorization_code"),
					ghttp.VerifyHeaderKV("authorization", "bearer client-bearer-token"),
					ghttp.RespondWith(http.StatusFound, "", http.Header{
						"Location": []string{"https://uaa.example.com/login?code=abc123"},
					}),
				))
			})

			It("gets the access code from the token endpoint", func() {
				runCommand()

				Ω(authRepo.RefreshTokenCalled).To(BeTrue())
				Ω(fakeUAA.ReceivedRequests()).To(HaveLen(1))
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"abc123"},
				))
			})

			It("dumps all the http requests and responses for logging", func() {
				var stdout *bytes.Buffer
				stdout = bytes.NewBuffer([]byte{})
				trace.SetStdout(stdout)

				trace.NewLogger("true")
				runCommand()

				result, err := ioutil.ReadAll(stdout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(ContainSubstring("REQUEST"))
				Expect(result).To(ContainSubstring("RESPONSE"))
			})

			It("returns an error when the uaa certificate is not valid and certificate validation is enabled", func() {
				configRepo.SetSSLDisabled(false)

				runCommand()

				Ω(authRepo.RefreshTokenCalled).To(BeTrue())
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"signed by unknown authority"},
				))
			})

			It("returns an error when the endpoint url cannot be parsed", func() {
				configRepo.SetUaaEndpoint(":goober#swallow?yak")

				runCommand()
				Ω(fakeUAA.ReceivedRequests()).To(HaveLen(0))
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"Error getting AuthenticationEndpoint"},
				))
			})

			It("returns an error when the request to the authorization server fails", func() {
				configRepo.SetUaaEndpoint("http://0.0.0.0") //invalid address

				runCommand()
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"Error requesting one time code from server"},
				))
			})

			It("returns an error when the authorization server does not redirect", func() {
				fakeUAA.RouteToHandler("GET", "/oauth/authorize", ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/oauth/authorize"),
					ghttp.RespondWith(http.StatusOK, ""),
				))

				runCommand()
				Ω(fakeUAA.ReceivedRequests()).To(HaveLen(1))
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"Authorization server did not redirect with one time code"},
				))
			})

			It("returns an error when the redirect URL does not contain a code", func() {
				fakeUAA.RouteToHandler("GET", "/oauth/authorize", ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/oauth/authorize"),
					ghttp.RespondWith(http.StatusFound, "", http.Header{
						"Location": []string{"https://uaa.example.com/login"},
					}),
				))

				runCommand()
				Ω(fakeUAA.ReceivedRequests()).To(HaveLen(1))
				Ω(ui.Outputs).To(ContainSubstrings(
					[]string{"Unable to acquire one time code from authorization response"},
				))
			})
		})

	})
})
