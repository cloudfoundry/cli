package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/api/uaa/uaaversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("auth Command", func() {
	var (
		cmd        AuthCommand
		testUI     *ui.UI
		fakeActor  *v7fakes.FakeActor
		fakeConfig *commandfakes.FakeConfig
		binaryName string
		err        error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v7fakes.FakeActor)
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = AuthCommand{
			BaseCommand: BaseCommand{
				UI:     testUI,
				Config: fakeConfig,
				Actor:  fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		err = cmd.Execute(nil)
	})

	When("--origin are set", func() {
		BeforeEach(func() {
			cmd.Origin = "some-origin"
		})

		When("the UAA is below the minimum API version", func() {
			BeforeEach(func() {
				fakeActor.UAAAPIVersionReturns(uaaversion.InvalidUAAClientVersion)
			})

			It("returns an API version error", func() {
				Expect(err).To(MatchError(translatableerror.MinimumUAAAPIVersionNotMetError{
					Command:        "Option '--origin'",
					MinimumVersion: uaaversion.MinUAAClientVersion,
				}))
			})
		})

		When("--client-credentials set", func() {
			BeforeEach(func() {
				cmd.ClientCredentials = true
				fakeActor.UAAAPIVersionReturns(uaaversion.MinUAAClientVersion)
			})

			It("returns an ArgumentCombinationError", func() {
				Expect(err).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--client-credentials", "--origin"},
				}))
			})
		})

		When("when the UAA is above the minimum API version", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.Username = "doesn't matter"
				cmd.RequiredArgs.Password = "doesn't matter"
				fakeActor.UAAAPIVersionReturns(uaaversion.MinUAAClientVersion)
			})

			It("authenticates with the values from the command line args", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
				_, origin, _ := fakeActor.AuthenticateArgsForCall(0)
				Expect(origin).To(Equal("some-origin"))
			})
		})
	})

	When("credentials are missing", func() {
		When("username and password are both missing", func() {
			It("raises an error", func() {
				Expect(err).To(MatchError(translatableerror.MissingCredentialsError{
					MissingUsername: true,
					MissingPassword: true,
				}))
			})
		})

		When("username is missing", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.Password = "mypassword"
			})

			It("raises an error", func() {
				Expect(err).To(MatchError(translatableerror.MissingCredentialsError{
					MissingUsername: true,
				}))
			})
		})

		When("password is missing", func() {
			BeforeEach(func() {
				cmd.RequiredArgs.Username = "myuser"
			})

			It("raises an error", func() {
				Expect(err).To(MatchError(translatableerror.MissingCredentialsError{
					MissingPassword: true,
				}))
			})
		})
	})

	When("there is an auth error", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.Username = "foo"
			cmd.RequiredArgs.Password = "bar"

			fakeConfig.TargetReturns("some-api-target")
			fakeActor.AuthenticateReturns(uaa.UnauthorizedError{Message: "some message"})
		})

		It("returns a BadCredentialsError", func() {
			Expect(err).To(MatchError(uaa.UnauthorizedError{Message: "some message"}))
		})
	})

	When("there is an account locked error", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.Username = "foo"
			cmd.RequiredArgs.Password = "bar"

			fakeConfig.TargetReturns("some-api-target")
			fakeActor.AuthenticateReturns(uaa.AccountLockedError{Message: "some message"})
		})

		It("returns a BadCredentialsError", func() {
			Expect(err).To(MatchError(uaa.AccountLockedError{Message: "some message"}))
		})
	})

	When("there is a non-auth error", func() {
		var expectedError error

		BeforeEach(func() {
			cmd.RequiredArgs.Username = "foo"
			cmd.RequiredArgs.Password = "bar"

			fakeConfig.TargetReturns("some-api-target")
			expectedError = errors.New("my humps")

			fakeActor.AuthenticateReturns(expectedError)
		})

		It("returns the error", func() {
			Expect(err).To(MatchError(expectedError))
		})
	})

	When("there are no input errors", func() {
		var (
			testID     string
			testSecret string
		)

		BeforeEach(func() {
			testID = "hello"
			testSecret = "goodbye"

			fakeConfig.TargetReturns("some-api-target")
		})

		When("--client-credentials is set", func() {
			BeforeEach(func() {
				cmd.ClientCredentials = true
				cmd.RequiredArgs.Username = testID
				cmd.RequiredArgs.Password = testSecret
			})

			It("outputs API target information and clears the targeted org and space", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("API endpoint: %s", fakeConfig.Target()))
				Expect(testUI.Out).To(Say(`Authenticating\.\.\.`))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).To(Say("Use '%s target' to view or set your target org and space", binaryName))

				Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
				credentials, origin, grantType := fakeActor.AuthenticateArgsForCall(0)
				ID := credentials["client_id"]
				secret := credentials["client_secret"]
				Expect(ID).To(Equal(testID))
				Expect(secret).To(Equal(testSecret))
				Expect(origin).To(BeEmpty())
				Expect(grantType).To(Equal(constant.GrantTypeClientCredentials))
			})
		})

		When("--client-credentials is not set", func() {
			When("username and password are only provided as arguments", func() {
				BeforeEach(func() {
					cmd.RequiredArgs.Username = testID
					cmd.RequiredArgs.Password = testSecret
				})

				It("outputs API target information and clears the targeted org and space", func() {
					Expect(err).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("API endpoint: %s", fakeConfig.Target()))
					Expect(testUI.Out).To(Say(`Authenticating\.\.\.`))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Use '%s target' to view or set your target org and space", binaryName))

					Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
					credentials, origin, grantType := fakeActor.AuthenticateArgsForCall(0)
					username := credentials["username"]
					password := credentials["password"]
					Expect(username).To(Equal(testID))
					Expect(password).To(Equal(testSecret))
					Expect(origin).To(BeEmpty())
					Expect(grantType).To(Equal(constant.GrantTypePassword))
				})
			})

			When("the username and password are provided in env variables", func() {
				var (
					envUsername string
					envPassword string
				)

				BeforeEach(func() {
					envUsername = "banana"
					envPassword = "potato"
					fakeConfig.CFUsernameReturns(envUsername)
					fakeConfig.CFPasswordReturns(envPassword)
				})

				When("username and password are not also provided as arguments", func() {
					It("authenticates with the values in the env variables", func() {
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
						credentials, origin, _ := fakeActor.AuthenticateArgsForCall(0)
						username := credentials["username"]
						password := credentials["password"]
						Expect(username).To(Equal(envUsername))
						Expect(password).To(Equal(envPassword))
						Expect(origin).To(BeEmpty())
					})
				})

				When("username and password are also provided as arguments", func() {
					BeforeEach(func() {
						cmd.RequiredArgs.Username = testID
						cmd.RequiredArgs.Password = testSecret
					})

					It("authenticates with the values from the command line args", func() {
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
						credentials, origin, _ := fakeActor.AuthenticateArgsForCall(0)
						username := credentials["username"]
						password := credentials["password"]
						Expect(username).To(Equal(testID))
						Expect(password).To(Equal(testSecret))
						Expect(origin).To(BeEmpty())
					})
				})
			})
		})

		When("client id or client secret are in the config.json", func() {
			BeforeEach(func() {
				fakeConfig.UAAOAuthClientReturns("AClientsId")
			})
			When("using --client-credentials", func() {
				BeforeEach(func() {
					cmd.ClientCredentials = true
					cmd.RequiredArgs.Username = "some-client-id"
					cmd.RequiredArgs.Password = "some-client-secret"

				})
				It("does not output a deprecation warning", func() {
					Expect(testUI.Err).ToNot(Say("Deprecation warning"))
				})
			})
			When("logging in as a user", func() {
				BeforeEach(func() {
					cmd.ClientCredentials = false
					cmd.RequiredArgs.Username = "some-username"
					cmd.RequiredArgs.Password = "some-password"
				})
				It("outputs a deprecation warning", func() {
					Expect(testUI.Err).To(Say("Deprecation warning: Manually writing your client credentials to the config.json is deprecated and will be removed in the future. For similar functionality, please use the `cf auth --client-credentials` command instead."))
				})
			})
		})
	})

	When("already logged in via client credentials", func() {
		BeforeEach(func() {
			fakeConfig.UAAGrantTypeReturns("client_credentials")
		})

		When("authenticating as a user", func() {
			BeforeEach(func() {
				cmd.ClientCredentials = false
				cmd.RequiredArgs.Username = "some-username"
				cmd.RequiredArgs.Password = "some-password"
			})

			It("returns an already logged in error", func() {
				Expect(err).To(MatchError("Service account currently logged in. Use 'cf logout' to log out service account and try again."))
				Expect(fakeConfig.UAAGrantTypeCallCount()).To(Equal(1))
			})

			It("the returned error is translatable", func() {
				Expect(err).To(MatchError(translatableerror.PasswordGrantTypeLogoutRequiredError{}))
			})
		})

		When("authenticating as a client", func() {
			BeforeEach(func() {
				cmd.ClientCredentials = true
				cmd.RequiredArgs.Username = "some-client-id"
				cmd.RequiredArgs.Password = "some-client-secret"
			})
			It("does not error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeConfig.UAAGrantTypeCallCount()).To(Equal(0))
			})
		})
	})
})
