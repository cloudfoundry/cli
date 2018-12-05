package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/api/uaa/uaaversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("auth Command", func() {
	var (
		cmd        AuthCommand
		testUI     *ui.UI
		fakeActor  *v6fakes.FakeAuthActor
		fakeConfig *commandfakes.FakeConfig
		binaryName string
		err        error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v6fakes.FakeAuthActor)
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = AuthCommand{
			UI:     testUI,
			Config: fakeConfig,
			Actor:  fakeActor,
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
				fakeActor.UAAAPIVersionReturns(uaaversion.MinUAAClientVersion)
			})

			It("returns an API version error", func() {
				Expect(err).To(MatchError(translatableerror.MinimumUAAAPIVersionNotMetError{
					Command:        "Option '--origin'",
					MinimumVersion: uaaversion.MinVersionOrigin,
				}))
			})
		})

		When("--client-credentials set", func() {
			BeforeEach(func() {
				cmd.ClientCredentials = true
				fakeActor.UAAAPIVersionReturns(uaaversion.MinVersionOrigin)
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
				fakeActor.UAAAPIVersionReturns(uaaversion.MinVersionOrigin)
			})

			It("authenticates with the values from the command line args", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
				_, _, origin, _ := fakeActor.AuthenticateArgsForCall(0)
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

	Describe("it checks the CLI version", func() {
		var (
			apiVersion    string
			minCLIVersion string
			binaryVersion string
		)

		BeforeEach(func() {
			apiVersion = "1.2.3"
			fakeActor.CloudControllerAPIVersionReturns(apiVersion)
			minCLIVersion = "1.0.0"
			fakeConfig.MinCLIVersionReturns(minCLIVersion)

			cmd.RequiredArgs.Username = "user"
			cmd.RequiredArgs.Password = "password"
		})

		Context("the CLI version is older than the minimum version required by the API", func() {
			BeforeEach(func() {
				binaryVersion = "0.0.0"
				fakeConfig.BinaryVersionReturns(binaryVersion)
			})

			It("displays a recommendation to update the CLI version", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
			})
		})

		Context("the CLI version satisfies the API's minimum version requirements", func() {
			BeforeEach(func() {
				binaryVersion = "1.0.0"
				fakeConfig.BinaryVersionReturns(binaryVersion)
			})

			It("does not display a recommendation to update the CLI version", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(testUI.Err).ToNot(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
			})
		})

		When("the CLI version is invalid", func() {
			BeforeEach(func() {
				fakeConfig.BinaryVersionReturns("&#%")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("No Major.Minor.Patch elements found"))
			})
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
				ID, secret, origin, grantType := fakeActor.AuthenticateArgsForCall(0)
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
					username, password, origin, grantType := fakeActor.AuthenticateArgsForCall(0)
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
						username, password, origin, _ := fakeActor.AuthenticateArgsForCall(0)
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
						username, password, origin, _ := fakeActor.AuthenticateArgsForCall(0)
						Expect(username).To(Equal(testID))
						Expect(password).To(Equal(testSecret))
						Expect(origin).To(BeEmpty())
					})
				})
			})
		})
	})
})
