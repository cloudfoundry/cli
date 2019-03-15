package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("login Command", func() {
	var (
		cmd              LoginCommand
		testUI           *ui.UI
		fakeActor        *v6fakes.FakeLoginActor
		fakeChecker      *v6fakes.FakeVersionChecker
		fakeConfig       *commandfakes.FakeConfig
		fakeActorMaker   *v6fakes.FakeActorMaker
		fakeCheckerMaker *v6fakes.FakeCheckerMaker
		executeErr       error
		input            *Buffer
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v6fakes.FakeLoginActor)
		fakeActorMaker = new(v6fakes.FakeActorMaker)
		fakeActorMaker.NewActorReturns(fakeActor, nil)

		fakeChecker = new(v6fakes.FakeVersionChecker)
		fakeCheckerMaker = new(v6fakes.FakeCheckerMaker)
		fakeCheckerMaker.NewVersionCheckerReturns(fakeChecker, nil)

		cmd = LoginCommand{
			UI:           testUI,
			Actor:        fakeActor,
			ActorMaker:   fakeActorMaker,
			Config:       fakeConfig,
			CheckerMaker: fakeCheckerMaker,
		}
		cmd.APIEndpoint = ""
		fakeConfig.BinaryNameReturns("faceman")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the experimental login flag is not set", func() {
		It("returns an UnrefactoredCommandError", func() {
			Expect(executeErr).To(MatchError(translatableerror.UnrefactoredCommandError{}))
		})
	})

	When("the experimental login flag is set", func() {
		BeforeEach(func() {
			fakeConfig.ExperimentalLoginReturns(true)
		})

		It("displays a helpful warning", func() {
			Expect(testUI.Err).To(Say("Using experimental login command, some behavior may be different"))
		})

		Describe("API Endpoint", func() {
			BeforeEach(func() {
				fakeConfig.APIVersionReturns("3.4.5")
			})

			When("user provides the api endpoint using the -a flag", func() {
				BeforeEach(func() {
					cmd.APIEndpoint = "api.boshlite.com"
				})

				It("target the provided api endpoint", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("API endpoint: api.boshlite.com"))
					Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
					actualSettings := fakeActor.SetTargetArgsForCall(0)
					Expect(actualSettings.URL).To(Equal("https://api.boshlite.com"))
				})
			})

			When("user does not provide the api endpoint using the -a flag", func() {
				When("config has API endpoint already set", func() {
					BeforeEach(func() {
						fakeConfig.TargetReturns("api.fake.com")
					})

					It("does not prompt the user for an API endpoint", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say(`API endpoint:\s+api\.fake\.com \(API version: 3\.4\.5\)`))
						Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
					})
				})

				When("the user enters something at the prompt", func() {
					BeforeEach(func() {
						input.Write([]byte("api.boshlite.com\n"))
						cmd.APIEndpoint = ""
					})

					It("targets the API that the user inputted", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("API endpoint:"))
						Expect(testUI.Out).To(Say("api.boshlite.com\n"))
						Expect(testUI.Out).To(Say(`API endpoint:\s+api\.boshlite\.com \(API version: 3\.4\.5\)`))

						Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
						actualSettings := fakeActor.SetTargetArgsForCall(0)
						Expect(actualSettings.URL).To(Equal("https://api.boshlite.com"))
					})
				})

				When("the user inputs an empty API", func() {
					BeforeEach(func() {
						cmd.APIEndpoint = ""
						input.Write([]byte("\n\napi.boshlite.com\n"))
					})

					It("reprompts for the API", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("API endpoint:"))
						Expect(testUI.Out).To(Say("API endpoint:"))
						Expect(testUI.Out).To(Say("API endpoint:"))
						Expect(testUI.Out).To(Say("api.boshlite.com\n"))
						Expect(testUI.Out).To(Say(`API endpoint:\s+api\.boshlite\.com \(API version: 3\.4\.5\)`))
					})
				})
			})

			When("the endpoint has trailing slashes", func() {
				BeforeEach(func() {
					cmd.APIEndpoint = "api.boshlite.com////"
				})

				It("strips the backslashes before using the endpoint", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
					actualSettings := fakeActor.SetTargetArgsForCall(0)
					Expect(actualSettings.URL).To(Equal("https://api.boshlite.com"))

					Expect(testUI.Out).To(Say(`API endpoint:\s+api\.boshlite\.com \(API version: 3\.4\.5\)`))
				})
			})
		})

		Describe("username and password", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("https://some.random.endpoint")
			})

			When("the current grant type is client credentials", func() {
				BeforeEach(func() {
					fakeConfig.UAAGrantTypeReturns(string(constant.GrantTypeClientCredentials))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("Service account currently logged in. Use 'cf logout' to log out service account and try again."))
				})
			})

			When("the current grant type is password", func() {
				BeforeEach(func() {
					fakeConfig.UAAGrantTypeReturns(string(constant.GrantTypePassword))
				})

				It("fetches prompts from the UAA", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.GetLoginPromptsCallCount()).To(Equal(1))
				})

				When("fetching prompts succeeds", func() {
					When("one of the prompts has a username key and is text type", func() {
						BeforeEach(func() {
							fakeActor.GetLoginPromptsReturns(map[string]coreconfig.AuthPrompt{
								"username": {
									DisplayName: "Username",
									Type:        coreconfig.AuthPromptTypeText,
								},
							})
						})

						When("the username flag is set", func() {
							BeforeEach(func() {
								cmd.Username = "potatoface"
							})

							It("uses the provided value and does not prompt for the username", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(testUI.Out).NotTo(Say("Username:"))
								Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
								credentials, _, _ := fakeActor.AuthenticateArgsForCall(0)
								Expect(credentials["username"]).To(Equal("potatoface"))
							})
						})
					})

					When("one of the prompts has password key and is password type", func() {
						BeforeEach(func() {
							fakeActor.GetLoginPromptsReturns(map[string]coreconfig.AuthPrompt{
								"password": {
									DisplayName: "Your Password",
									Type:        coreconfig.AuthPromptTypePassword,
								},
							})
						})

						When("the password flag is set", func() {
							BeforeEach(func() {
								cmd.Password = "noprompto"
							})

							It("uses the provided value and does not prompt for the password", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(testUI.Out).NotTo(Say("Your Password:"))
								Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
								credentials, _, _ := fakeActor.AuthenticateArgsForCall(0)
								Expect(credentials["password"]).To(Equal("noprompto"))
							})

							When("the password is incorrect", func() {
								BeforeEach(func() {
									input.Write([]byte("other-password\n"))
									fakeActor.AuthenticateReturns(errors.New("bad creds"))
								})

								It("does not reuse the flag value for subsequent attempts", func() {
									credentials, _, _ := fakeActor.AuthenticateArgsForCall(1)
									Expect(credentials["password"]).To(Equal("other-password"))
								})
							})
						})
					})

					When("UAA prompts for the SSO passcode during non-SSO flow", func() {
						BeforeEach(func() {
							cmd.SSO = false
							cmd.Password = "some-password"
							fakeActor.GetLoginPromptsReturns(map[string]coreconfig.AuthPrompt{
								"password": {
									DisplayName: "Your Password",
									Type:        coreconfig.AuthPromptTypePassword,
								},
								"passcode": {
									DisplayName: "gimme your passcode",
									Type:        coreconfig.AuthPromptTypePassword,
								},
							})
						})

						It("does not prompt for the passcode", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).NotTo(Say("gimme your passcode"))
						})

						It("does not send the passcode", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							credentials, _, _ := fakeActor.AuthenticateArgsForCall(0)
							Expect(credentials).To(HaveKeyWithValue("password", "some-password"))
							Expect(credentials).NotTo(HaveKey("passcode"))
						})
					})

					When("multiple prompts of text and password type are returned", func() {
						BeforeEach(func() {
							fakeActor.GetLoginPromptsReturns(map[string]coreconfig.AuthPrompt{
								"account_number": {
									DisplayName: "Account Number",
									Type:        coreconfig.AuthPromptTypeText,
								},
								"username": {
									DisplayName: "Username",
									Type:        coreconfig.AuthPromptTypeText,
								},
								"passcode": {
									DisplayName: "It's a passcode, what you want it to be???",
									Type:        coreconfig.AuthPromptTypePassword,
								},
								"password": {
									DisplayName: "Your Password",
									Type:        coreconfig.AuthPromptTypePassword,
								},
								"supersecret": {
									DisplayName: "Meaning of Life",
									Type:        coreconfig.AuthPromptTypePassword,
								},
							})
						})

						When("no authentication flags are set", func() {
							BeforeEach(func() {
								input.Write([]byte("fakeman\nsomeaccount\nsomepassword\ngarbage\n"))
							})

							It("displays text prompts, starting with username, then password prompts, starting with password", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Username:"))
								Expect(testUI.Out).To(Say("fakeman"))
								Expect(testUI.Out).To(Say("Account Number:"))
								Expect(testUI.Out).To(Say("someaccount"))

								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).NotTo(Say("somepassword"))
								Expect(testUI.Out).To(Say("Meaning of Life:"))
								Expect(testUI.Out).NotTo(Say("garbage"))
							})

							It("authenticates with the responses", func() {
								Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
								credentials, _, grantType := fakeActor.AuthenticateArgsForCall(0)
								Expect(credentials["username"]).To(Equal("fakeman"))
								Expect(credentials["password"]).To(Equal("somepassword"))
								Expect(credentials["supersecret"]).To(Equal("garbage"))
								Expect(grantType).To(Equal(constant.GrantTypePassword))
							})
						})

						When("authenticating succeeds", func() {
							BeforeEach(func() {
								fakeConfig.CurrentUserNameReturns("happyman", nil)
							})

							It("displays OK and a status summary", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
								Expect(testUI.Out).To(Say(`User:\s+happyman`))

								Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
							})
						})

						When("authenticating fails", func() {
							BeforeEach(func() {
								fakeActor.AuthenticateReturns(errors.New("something died"))
							})

							It("prints the error message three times", func() {
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("Meaning of Life:"))
								Expect(testUI.Err).To(Say("something died"))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("Meaning of Life:"))
								Expect(testUI.Err).To(Say("something died"))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("Meaning of Life:"))
								Expect(testUI.Err).To(Say("something died"))
							})

							It("returns an error indicating that it could not authenticate", func() {
								Expect(executeErr).To(MatchError("Unable to authenticate."))
							})

							It("displays a status summary", func() {
								Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
								Expect(testUI.Out).To(Say(`Not logged in. Use '%s login' to log in.`, cmd.Config.BinaryName()))
							})

						})

						When("authenticating fails with a bad credentials error", func() {
							BeforeEach(func() {
								fakeActor.AuthenticateReturns(uaa.UnauthorizedError{Message: "Bad credentials"})
							})

							It("converts the error before printing it", func() {
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("Meaning of Life:"))
								Expect(testUI.Err).To(Say("Credentials were rejected, please try again."))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("Meaning of Life:"))
								Expect(testUI.Err).To(Say("Credentials were rejected, please try again."))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("Meaning of Life:"))
								Expect(testUI.Err).To(Say("Credentials were rejected, please try again."))
							})
						})
					})
				})
			})
		})

		Describe("Minimum CLI version ", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("whatever.com")

				fakeChecker.MinCLIVersionReturns("9000.0.0")
			})

			It("sets the minimum CLI version in the config", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeConfig.SetMinCLIVersionCallCount()).To(Equal(1))
				Expect(fakeConfig.SetMinCLIVersionArgsForCall(0)).To(Equal("9000.0.0"))
			})

			When("The current version is below the minimum supported", func() {
				BeforeEach(func() {
					fakeChecker.CloudControllerAPIVersionReturns("2.123.0")
					fakeConfig.BinaryVersionReturns("1.2.3")
					fakeConfig.MinCLIVersionReturns("9000.0.0")
				})

				It("displays a warning", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("Cloud Foundry API version 2.123.0 requires CLI version 9000.0.0. You are currently on version 1.2.3. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads"))
				})

				Context("ordering of output", func() {
					BeforeEach(func() {
						outAndErr := NewBuffer()
						testUI.Out = outAndErr
						testUI.Err = outAndErr
					})

					It("displays the warning after all prompts but before the summary ", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(testUI.Out).To(Say(`Authenticating...`))
						Expect(testUI.Err).To(Say("Cloud Foundry API version 2.123.0 requires CLI version 9000.0.0. You are currently on version 1.2.3. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads"))
						Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
						Expect(testUI.Out).To(Say(`Not logged in. Use 'faceman login' to log in.`))
					})
				})
			})
		})
	})
})
