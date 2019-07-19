package v6_test

import (
	"errors"
	"fmt"
	"io"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("login Command", func() {
	var (
		binaryName       string
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
		binaryName = "some-executable"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = LoginCommand{
			UI:           testUI,
			Actor:        fakeActor,
			ActorMaker:   fakeActorMaker,
			Config:       fakeConfig,
			CheckerMaker: fakeCheckerMaker,
		}
		cmd.APIEndpoint = ""
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
					Expect(testUI.Out).To(Say("API endpoint: api.boshlite.com\n\n"))
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

				When("client secret in the configuration is present", func() {
					It("should not display a warning", func() {
						Expect(testUI.Err).NotTo(Say("Deprecation warning:"))
					})
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

							When("there have been too many failed login attempts", func() {
								BeforeEach(func() {
									input.Write([]byte("other-password\n"))
									fakeActor.AuthenticateReturns(
										uaa.AccountLockedError{
											Message: "Your account has been locked because of too many failed attempts to login.",
										},
									)
								})

								It("does not reuse the flag value for subsequent attempts", func() {
									Expect(fakeActor.AuthenticateCallCount()).To(Equal(1), "called Authenticate again after lockout")
									Expect(testUI.Err).To(Say("Your account has been locked because of too many failed attempts to login."))
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
									DisplayName: "MFA Code",
									Type:        coreconfig.AuthPromptTypePassword,
								},
							})
						})

						When("no authentication flags are set", func() {
							BeforeEach(func() {
								input.Write([]byte("faker\nsomeaccount\nsomepassword\ngarbage\n"))
							})

							It("displays text prompts, starting with username, then password prompts, starting with password", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say("Username:"))
								Expect(testUI.Out).To(Say("faker"))

								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say("Account Number:"))
								Expect(testUI.Out).To(Say("someaccount"))

								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).NotTo(Say("somepassword"))

								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say("MFA Code:"))
								Expect(testUI.Out).NotTo(Say("garbage"))
							})

							It("authenticates with the responses", func() {
								Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
								credentials, _, grantType := fakeActor.AuthenticateArgsForCall(0)
								Expect(credentials["username"]).To(Equal("faker"))
								Expect(credentials["password"]).To(Equal("somepassword"))
								Expect(credentials["supersecret"]).To(Equal("garbage"))
								Expect(grantType).To(Equal(constant.GrantTypePassword))
							})
						})

						When("an error occurs prompting for the username", func() {
							var fakeUI *commandfakes.FakeUI

							BeforeEach(func() {
								fakeUI = new(commandfakes.FakeUI)
								fakeUI.DisplayTextPromptReturns("", errors.New("some-error"))
								cmd = LoginCommand{
									UI:           fakeUI,
									Actor:        fakeActor,
									ActorMaker:   fakeActorMaker,
									Config:       fakeConfig,
									CheckerMaker: fakeCheckerMaker,
								}
							})

							It("stops prompting after the first prompt", func() {
								Expect(fakeUI.DisplayTextPromptCallCount()).To(Equal(1))
							})

							It("errors", func() {
								Expect(executeErr).To(MatchError("Unable to authenticate."))
							})
						})

						When("an error occurs in an additional text prompt after username", func() {
							var fakeUI *commandfakes.FakeUI

							BeforeEach(func() {
								fakeUI = new(commandfakes.FakeUI)
								fakeUI.DisplayTextPromptReturnsOnCall(0, "some-name", nil)
								fakeUI.DisplayTextPromptReturnsOnCall(1, "", errors.New("some-error"))
								cmd = LoginCommand{
									UI:           fakeUI,
									Actor:        fakeActor,
									ActorMaker:   fakeActorMaker,
									Config:       fakeConfig,
									CheckerMaker: fakeCheckerMaker,
								}
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError("Unable to authenticate."))
							})
						})

						When("an error occurs prompting for the password", func() {
							var fakeUI *commandfakes.FakeUI

							BeforeEach(func() {
								fakeUI = new(commandfakes.FakeUI)
								fakeUI.DisplayPasswordPromptReturns("", errors.New("some-error"))
								cmd = LoginCommand{
									UI:           fakeUI,
									Actor:        fakeActor,
									ActorMaker:   fakeActorMaker,
									Config:       fakeConfig,
									CheckerMaker: fakeCheckerMaker,
								}
							})

							It("stops prompting after the first prompt", func() {
								Expect(fakeUI.DisplayPasswordPromptCallCount()).To(Equal(1))
							})

							It("errors", func() {
								Expect(executeErr).To(MatchError("Unable to authenticate."))
							})
						})

						When("an error occurs prompting for prompts of type password that are not the 'password'", func() {
							var fakeUI *commandfakes.FakeUI

							BeforeEach(func() {
								fakeUI = new(commandfakes.FakeUI)
								fakeUI.DisplayPasswordPromptReturnsOnCall(0, "some-password", nil)
								fakeUI.DisplayPasswordPromptReturnsOnCall(1, "", errors.New("some-error"))

								cmd = LoginCommand{
									UI:           fakeUI,
									Actor:        fakeActor,
									ActorMaker:   fakeActorMaker,
									Config:       fakeConfig,
									CheckerMaker: fakeCheckerMaker,
								}
							})

							It("stops prompting after the second prompt", func() {
								Expect(fakeUI.DisplayPasswordPromptCallCount()).To(Equal(2))
							})

							It("errors", func() {
								Expect(executeErr).To(MatchError("Unable to authenticate."))
							})
						})

						When("authenticating succeeds", func() {
							BeforeEach(func() {
								fakeConfig.CurrentUserNameReturns("potatoface", nil)
								input.Write([]byte("faker\nsomeaccount\nsomepassword\ngarbage\n"))
							})

							It("displays OK and a status summary", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
								Expect(testUI.Out).To(Say(`User:\s+potatoface`))

								Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
							})
						})

						When("authenticating fails", func() {
							BeforeEach(func() {
								fakeActor.AuthenticateReturns(errors.New("something died"))
								input.Write([]byte("faker\nsomeaccount\nsomepassword\ngarbage\nfaker\nsomeaccount\nsomepassword\ngarbage\nfaker\nsomeaccount\nsomepassword\ngarbage\n"))
							})

							It("prints the error message three times", func() {
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("MFA Code:"))
								Expect(testUI.Err).To(Say("something died"))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("MFA Code:"))
								Expect(testUI.Err).To(Say("something died"))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("MFA Code:"))
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
								input.Write([]byte("faker\nsomeaccount\nsomepassword\ngarbage\nfaker\nsomeaccount\nsomepassword\ngarbage\nfaker\nsomeaccount\nsomepassword\ngarbage\n"))
							})

							It("converts the error before printing it", func() {
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("MFA Code:"))
								Expect(testUI.Err).To(Say("Credentials were rejected, please try again."))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("MFA Code:"))
								Expect(testUI.Err).To(Say("Credentials were rejected, please try again."))
								Expect(testUI.Out).To(Say("Your Password:"))
								Expect(testUI.Out).To(Say("MFA Code:"))
								Expect(testUI.Err).To(Say("Credentials were rejected, please try again."))
							})
						})
					})

					When("custom client ID and client secret are set in the config file", func() {
						BeforeEach(func() {
							fakeConfig.UAAOAuthClientReturns("some-other-client-id")
							fakeConfig.UAAOAuthClientSecretReturns("some-secret")
						})

						It("prints a deprecation warning", func() {
							deprecationMessage := "Deprecation warning: Manually writing your client credentials to the config.json is deprecated and will be removed in the future. For similar functionality, please use the `cf auth --client-credentials` command instead."
							Expect(testUI.Err).To(Say(deprecationMessage))
						})

						It("still attempts to log in", func() {
							Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
						})
					})
				})
			})
		})

		Describe("SSO Passcode", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("whatever.com")

				input.Write([]byte("some-passcode\n"))
				fakeActor.GetLoginPromptsReturns(map[string]coreconfig.AuthPrompt{
					"passcode": {
						DisplayName: "some-sso-prompt",
						Type:        coreconfig.AuthPromptTypePassword,
					},
				})

				fakeConfig.CurrentUserNameReturns("potatoface", nil)
			})

			When("--sso flag is set", func() {
				BeforeEach(func() {
					cmd.SSO = true
				})

				It("prompts the user for SSO passcode", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(fakeActor.GetLoginPromptsCallCount()).To(Equal(1))
					Expect(testUI.Out).To(Say("some-sso-prompt:"))
				})

				It("authenticates with the inputted code", func() {
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
					Expect(testUI.Out).To(Say(`User:\s+potatoface`))

					Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
					credentials, origin, grantType := fakeActor.AuthenticateArgsForCall(0)
					Expect(credentials["passcode"]).To(Equal("some-passcode"))
					Expect(origin).To(BeEmpty())
					Expect(grantType).To(Equal(constant.GrantTypePassword))
				})

				When("an error occurs prompting for the code", func() {
					var fakeUI *commandfakes.FakeUI

					BeforeEach(func() {
						fakeUI = new(commandfakes.FakeUI)
						fakeUI.DisplayPasswordPromptReturns("", errors.New("some-error"))
						cmd = LoginCommand{
							UI:           fakeUI,
							Actor:        fakeActor,
							ActorMaker:   fakeActorMaker,
							Config:       fakeConfig,
							CheckerMaker: fakeCheckerMaker,
							SSO:          true,
						}
					})

					It("errors", func() {
						Expect(fakeUI.DisplayPasswordPromptCallCount()).To(Equal(1))
						Expect(executeErr).To(MatchError("Unable to authenticate."))
					})
				})
			})

			When("the --sso-passcode flag is set", func() {
				BeforeEach(func() {
					cmd.SSOPasscode = "a-passcode"
				})

				It("does not prompt the user for SSO passcode", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).ToNot(Say("some-sso-prompt:"))
				})

				It("uses the flag value to authenticate", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
					credentials, origin, grantType := fakeActor.AuthenticateArgsForCall(0)
					Expect(credentials["passcode"]).To(Equal("a-passcode"))
					Expect(origin).To(BeEmpty())
					Expect(grantType).To(Equal(constant.GrantTypePassword))
				})

				It("displays a summary with user information", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
					Expect(testUI.Out).To(Say(`User:\s+potatoface`))
				})

				When("an incorrect passcode is inputted", func() {
					BeforeEach(func() {
						cmd.SSOPasscode = "some-garbage"
						fakeActor.AuthenticateReturns(uaa.UnauthorizedError{
							Message: "Bad credentials",
						})
						fakeConfig.CurrentUserNameReturns("", nil)
						input.Write([]byte("some-passcode\n"))
					})

					It("re-prompts two more times", func() {
						Expect(testUI.Out).To(Say("some-sso-prompt:"))
						Expect(testUI.Out).To(Say(`Authenticating\.\.\.`))
						Expect(testUI.Err).To(Say("Credentials were rejected, please try again."))
						Expect(testUI.Out).To(Say("some-sso-prompt:"))
						Expect(testUI.Out).To(Say(`Authenticating\.\.\.`))
						Expect(testUI.Err).To(Say("Credentials were rejected, please try again."))
					})

					It("returns an error message", func() {
						Expect(executeErr).To(MatchError("Unable to authenticate."))
					})

					It("does not include user information in the summary", func() {
						Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
						Expect(testUI.Out).To(Say(`Not logged in. Use '%s login' to log in.`, cmd.Config.BinaryName()))
					})
				})
			})

			When("both --sso and --sso-passcode flags are set", func() {
				BeforeEach(func() {
					cmd.SSO = true
					cmd.SSOPasscode = "a-passcode"
				})

				It("returns an error message", func() {
					Expect(fakeActor.AuthenticateCallCount()).To(Equal(0))
					Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{Args: []string{"--sso-passcode", "--sso"}}))
				})
			})
		})

		Describe("Minimum CLI version", func() {
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
						Expect(testUI.Out).To(Say(`Not logged in. Use '%s login' to log in.`, binaryName))
					})
				})
			})
		})

		Describe("Targeting Org", func() {
			BeforeEach(func() {
				cmd.APIEndpoint = "example.com"
				cmd.Username = "some-user"
				cmd.Password = "some-password"
				fakeConfig.APIVersionReturns("3.4.5")
				fakeConfig.CurrentUserNameReturns("some-user", nil)
			})

			When("-o was passed", func() {
				BeforeEach(func() {
					cmd.Organization = "some-org"
				})

				It("fetches the specified organization", func() {
					Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
					Expect(fakeActor.GetOrganizationsCallCount()).To(Equal(0))
					Expect(fakeActor.GetOrganizationByNameArgsForCall(0)).To(Equal("some-org"))
				})

				When("fetching the organization succeeds", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationByNameReturns(
							v3action.Organization{Name: "some-org", GUID: "some-guid"},
							v3action.Warnings{"some-warning-1", "some-warning-2"},
							nil)
						fakeConfig.TargetedOrganizationNameReturns("some-org")
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("sets the targeted organization in the config", func() {
						Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
						orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
						Expect(orgGUID).To(Equal("some-guid"))
						Expect(orgName).To(Equal("some-org"))
					})

					It("reports to the user that the org is targeted", func() {
						Expect(testUI.Out).To(Say("API endpoint:   example.com \\(API version: 3.4.5\\)"))
						Expect(testUI.Out).To(Say("User:           some-user"))
						Expect(testUI.Out).To(Say("Org:            some-org"))
					})
				})

				When("fetching  the organization fails", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationByNameReturns(
							v3action.Organization{},
							v3action.Warnings{"some-warning-1", "some-warning-2"},
							errors.New("org-not-found"),
						)
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})

					It("does not set the targeted org", func() {
						Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(0))
					})
				})
			})

			When("-o was not passed", func() {
				BeforeEach(func() {
					cmd.APIEndpoint = "example.com"
					cmd.Username = "some-user"
					cmd.Password = "some-password"
					fakeActor.GetOrganizationsReturns(
						[]v3action.Organization{},
						v3action.Warnings{"some-org-warning-1", "some-org-warning-2"},
						nil,
					)
				})

				It("fetches the available organizations", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.GetOrganizationsCallCount()).To(Equal(1))
				})

				It("prints all warnings", func() {
					Expect(testUI.Err).To(Say("some-org-warning-1"))
					Expect(testUI.Err).To(Say("some-org-warning-2"))
				})

				When("fetching the organizations succeeds", func() {
					BeforeEach(func() {
						fakeConfig.CurrentUserNameReturns("some-user", nil)
					})

					When("no org exists", func() {
						It("does not prompt the user to select an org", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).ToNot(Say("Select an org:"))
							Expect(testUI.Out).ToNot(Say(`Org \(enter to skip\):`))
						})

						It("displays how to target an org and space", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("API endpoint:   example.com \\(API version: 3.4.5\\)"))
							Expect(testUI.Out).To(Say("User:           some-user"))
							Expect(testUI.Out).To(Say("No org or space targeted, use '%s target -o ORG -s SPACE'", binaryName))
						})
					})

					When("only one org exists", func() {
						BeforeEach(func() {
							fakeActor.GetOrganizationsReturns(
								[]v3action.Organization{v3action.Organization{
									GUID: "some-org-guid",
									Name: "some-org-name",
								}},
								v3action.Warnings{"some-org-warning-1", "some-org-warning-2"},
								nil,
							)
						})

						It("targets that org", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
							orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
							Expect(orgGUID).To(Equal("some-org-guid"))
							Expect(orgName).To(Equal("some-org-name"))
						})
					})

					When("more than one but fewer than 50 orgs exists", func() {
						BeforeEach(func() {
							fakeActor.GetOrganizationsReturns(
								[]v3action.Organization{
									v3action.Organization{
										GUID: "some-org-guid3",
										Name: "1234",
									},
									v3action.Organization{
										GUID: "some-org-guid1",
										Name: "some-org-name1",
									},
									v3action.Organization{
										GUID: "some-org-guid2",
										Name: "some-org-name2",
									},
								},
								v3action.Warnings{"some-org-warning-1", "some-org-warning-2"},
								nil,
							)
						})

						When("the user selects an org by list position", func() {
							When("the position is valid", func() {
								BeforeEach(func() {
									input.Write([]byte("2\n"))
								})

								It("prompts the user to select an org", func() {
									Expect(testUI.Out).To(Say("Select an org:"))
									Expect(testUI.Out).To(Say("1. 1234"))
									Expect(testUI.Out).To(Say("2. some-org-name1"))
									Expect(testUI.Out).To(Say("3. some-org-name2"))
									Expect(testUI.Out).To(Say(`Org \(enter to skip\):`))
									Expect(executeErr).ToNot(HaveOccurred())
								})

								It("targets that org", func() {
									Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
									orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
									Expect(orgGUID).To(Equal("some-org-guid1"))
									Expect(orgName).To(Equal("some-org-name1"))
								})
							})

							When("the position is invalid", func() {
								BeforeEach(func() {
									input.Write([]byte("4\n"))
								})

								It("reprompts the user", func() {
									Expect(testUI.Out).To(Say("Select an org:"))
									Expect(testUI.Out).To(Say("1. 1234"))
									Expect(testUI.Out).To(Say("2. some-org-name1"))
									Expect(testUI.Out).To(Say("3. some-org-name2"))
									Expect(testUI.Out).To(Say(`Org \(enter to skip\):`))
									Expect(testUI.Out).To(Say("Select an org:"))
									Expect(testUI.Out).To(Say("1. 1234"))
									Expect(testUI.Out).To(Say("2. some-org-name1"))
									Expect(testUI.Out).To(Say("3. some-org-name2"))
									Expect(testUI.Out).To(Say(`Org \(enter to skip\):`))
								})
							})
						})

						When("the user selects an org by name", func() {
							When("the list contains that org", func() {
								BeforeEach(func() {
									input.Write([]byte("some-org-name2\n"))
								})

								It("prompts the user to select an org", func() {
									Expect(testUI.Out).To(Say("Select an org:"))
									Expect(testUI.Out).To(Say("1. 1234"))
									Expect(testUI.Out).To(Say("2. some-org-name1"))
									Expect(testUI.Out).To(Say("3. some-org-name2"))
									Expect(testUI.Out).To(Say(`Org \(enter to skip\):`))
									Expect(executeErr).ToNot(HaveOccurred())
								})

								It("targets that org", func() {
									Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
									orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
									Expect(orgGUID).To(Equal("some-org-guid2"))
									Expect(orgName).To(Equal("some-org-name2"))
								})
							})

							When("the org is not in the list", func() {
								BeforeEach(func() {
									input.Write([]byte("invalid-org\n"))
								})

								It("returns an error", func() {
									Expect(executeErr).To(MatchError(translatableerror.OrganizationNotFoundError{Name: "invalid-org"}))
								})

								It("does not target the org", func() {
									Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(0))
								})
							})
						})

						When("the user exits the prompt early", func() {
							var fakeUI *commandfakes.FakeUI

							BeforeEach(func() {
								fakeUI = new(commandfakes.FakeUI)
								cmd.UI = fakeUI
							})

							When("the prompt returns with an EOF", func() {
								BeforeEach(func() {
									fakeUI.DisplayTextMenuReturns("", io.EOF)
								})

								It("selects no org and returns no error", func() {
									Expect(executeErr).ToNot(HaveOccurred())
									Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(0))
								})
							})
						})

					})

					When("more than 50 orgs exist", func() {
						BeforeEach(func() {
							orgs := make([]v3action.Organization, 51)
							for i := range orgs {
								orgs[i].Name = fmt.Sprintf("org%d", i+1)
								orgs[i].GUID = fmt.Sprintf("org-guid%d", i+1)
							}

							fakeActor.GetOrganizationsReturns(
								orgs,
								v3action.Warnings{"some-org-warning-1", "some-org-warning-2"},
								nil,
							)
						})

						When("the user selects an org by name", func() {
							When("the list contains that org", func() {
								BeforeEach(func() {
									input.Write([]byte("org37\n"))
								})

								It("prompts the user to select an org", func() {
									Expect(testUI.Out).To(Say("There are too many options to display; please type in the name."))
									Expect(testUI.Out).To(Say(`Org \(enter to skip\):`))
									Expect(executeErr).ToNot(HaveOccurred())
								})

								It("targets that org", func() {
									Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
									orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
									Expect(orgGUID).To(Equal("org-guid37"))
									Expect(orgName).To(Equal("org37"))
								})
							})

							When("the org is not in the list", func() {
								BeforeEach(func() {
									input.Write([]byte("invalid-org\n"))
								})

								It("returns an error", func() {
									Expect(executeErr).To(MatchError(translatableerror.OrganizationNotFoundError{Name: "invalid-org"}))
								})

								It("does not target the org", func() {
									Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(0))
								})
							})
						})

					})
				})

				When("fetching the organizations fails", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationsReturns(
							[]v3action.Organization{},
							v3action.Warnings{"some-warning-1", "some-warning-2"},
							errors.New("api call failed"),
						)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("api call failed"))
					})

					It("prints all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})
				})
			})
		})

		Describe("Targeting Space", func() {
			BeforeEach(func() {
				cmd.APIEndpoint = "example.com"
				cmd.Username = "some-user"
				cmd.Password = "some-password"
				fakeConfig.APIVersionReturns("3.4.5")
				fakeConfig.CurrentUserNameReturns("some-user", nil)
			})

			When("an org has been successfully targeted", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						GUID: "targeted-org-guid",
						Name: "targeted-org-name"},
					)
					fakeConfig.TargetedOrganizationNameReturns("targeted-org-name")
				})

				When("-s was passed", func() {
					BeforeEach(func() {
						cmd.Space = "some-space"
					})

					When("the specified space exists", func() {
						BeforeEach(func() {
							fakeActor.GetSpaceByNameAndOrganizationReturns(
								v3action.Space{
									Name: "some-space",
									GUID: "some-space-guid",
								},
								v3action.Warnings{"some-warning-1", "some-warning-2"},
								nil,
							)
						})

						It("targets that space", func() {
							Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))
							spaceGUID, spaceName, allowSSH := fakeConfig.SetSpaceInformationArgsForCall(0)
							Expect(spaceGUID).To(Equal("some-space-guid"))
							Expect(spaceName).To(Equal("some-space"))
							Expect(allowSSH).To(BeTrue())
						})

						It("prints all warnings", func() {
							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))
						})

						When("the space has been successfully targeted", func() {
							BeforeEach(func() {
								fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space"})
							})

							It("displays that the spacce has been targeted", func() {
								Expect(testUI.Out).To(Say(`Space:\s+some-space`))
							})
						})
					})

					When("the specified space does not exist or does not belong to the targeted org", func() {
						BeforeEach(func() {
							fakeActor.GetSpaceByNameAndOrganizationReturns(
								v3action.Space{},
								v3action.Warnings{"some-warning-1", "some-warning-2"},
								actionerror.SpaceNotFoundError{Name: "some-space"},
							)
						})

						It("returns an error", func() {
							Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: "some-space"}))
						})

						It("prints all warnings", func() {
							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))
						})

						It("reports that no space is targeted", func() {
							Expect(testUI.Out).To(Say(`Space:\s+No space targeted, use 'some-executable target -s SPACE'`))
						})
					})
				})

				When("-s was not passed", func() {
					When("no space exists", func() {
						BeforeEach(func() {
							fakeActor.GetOrganizationSpacesReturns(
								[]v3action.Space{},
								v3action.Warnings{},
								nil,
							)
						})
						It("does not prompt the user to select a space", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).ToNot(Say("Select a space:"))
							Expect(testUI.Out).ToNot(Say(`Space \(enter to skip\):`))
						})

						It("displays how to target a space", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("API endpoint:   example.com \\(API version: 3.4.5\\)"))
							Expect(testUI.Out).To(Say("User:           some-user"))
							Expect(testUI.Out).To(Say("No space targeted, use '%s target -s SPACE'", binaryName))
						})
					})

					When("only one space is available", func() {
						BeforeEach(func() {
							spaces := []v3action.Space{
								{
									GUID:             "some-space-guid",
									Name:             "some-space-name",
									OrganizationGUID: "targeted-org-guid",
								},
							}

							fakeActor.GetOrganizationSpacesReturns(
								spaces,
								v3action.Warnings{"some-warning-1", "some-warning-2"},
								nil,
							)

							fakeConfig.TargetedSpaceReturns(configv3.Space{
								GUID: "some-space-guid",
								Name: "some-space-name",
							})
						})

						It("targets this space", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(fakeActor.GetOrganizationSpacesCallCount()).To(Equal(1))
							Expect(fakeActor.GetOrganizationSpacesArgsForCall(0)).To(Equal("targeted-org-guid"))

							Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))

							firstArg, secondArg, _ := fakeConfig.SetSpaceInformationArgsForCall(0)
							Expect(firstArg).To(Equal("some-space-guid"))
							Expect(secondArg).To(Equal("some-space-name"))

							Expect(testUI.Err).To(Say("some-warning-1"))
							Expect(testUI.Err).To(Say("some-warning-2"))

							Expect(testUI.Out).To(Say(`Targeted space: some-space-name`))
							Expect(testUI.Out).To(Say(`Space:\s+some-space-name`))
							Expect(testUI.Out).NotTo(Say(`Space:\s+No space targeted, use 'some-executable target -s SPACE`))
						})
					})

					When("more than one space is available", func() {
						BeforeEach(func() {
							spaces := []v3action.Space{
								{
									GUID:             "some-space-guid",
									Name:             "some-space-name",
									OrganizationGUID: "targeted-org-guid",
								},
								{
									GUID:             "some-space-guid1",
									Name:             "some-space-name1",
									OrganizationGUID: "targeted-org-guid1",
								},
								{
									GUID:             "some-space-guid2",
									Name:             "some-space-name2",
									OrganizationGUID: "targeted-org-guid2",
								},
							}

							fakeActor.GetOrganizationSpacesReturns(
								spaces,
								v3action.Warnings{},
								nil,
							)
						})

						It("displays a numbered list of spaces", func() {
							Expect(testUI.Out).To(Say("Select a space:"))
							Expect(testUI.Out).To(Say("1. some-space-name"))
							Expect(testUI.Out).To(Say("2. some-space-name1"))
							Expect(testUI.Out).To(Say("3. some-space-name2"))
							Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
						})

						When("the user selects a space by list position", func() {
							When("the position is valid", func() {
								BeforeEach(func() {
									input.Write([]byte("2\n"))
								})

								It("targets that space", func() {
									Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))
									guid, name, allowSSH := fakeConfig.SetSpaceInformationArgsForCall(0)
									Expect(guid).To(Equal("some-space-guid1"))
									Expect(name).To(Equal("some-space-name1"))
									Expect(allowSSH).To(BeTrue())
									Expect(executeErr).NotTo(HaveOccurred())
								})
							})
							When("the position is invalid", func() {
								BeforeEach(func() {
									input.Write([]byte("100\n"))
								})

								It("reprompts the user", func() {
									Expect(testUI.Out).To(Say("Select a space:"))
									Expect(testUI.Out).To(Say("1. some-space-name"))
									Expect(testUI.Out).To(Say("2. some-space-name1"))
									Expect(testUI.Out).To(Say("3. some-space-name2"))
									Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
									Expect(testUI.Out).To(Say("1. some-space-name"))
									Expect(testUI.Out).To(Say("2. some-space-name1"))
									Expect(testUI.Out).To(Say("3. some-space-name2"))
									Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
								})

							})
						})

						When("the user selects a space by name", func() {
							When("the list contains that space", func() {
								BeforeEach(func() {
									input.Write([]byte("some-space-name2\n"))
								})

								It("prompts the user to select a space", func() {
									Expect(testUI.Out).To(Say("Select a space:"))
									Expect(testUI.Out).To(Say("1. some-space-name"))
									Expect(testUI.Out).To(Say("2. some-space-name1"))
									Expect(testUI.Out).To(Say("3. some-space-name2"))
									Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
									Expect(executeErr).ToNot(HaveOccurred())
								})

								It("targets that space", func() {
									Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))
									guid, name, allowSSH := fakeConfig.SetSpaceInformationArgsForCall(0)
									Expect(guid).To(Equal("some-space-guid2"))
									Expect(name).To(Equal("some-space-name2"))
									Expect(allowSSH).To(BeTrue())
									Expect(executeErr).NotTo(HaveOccurred())
								})
							})

							When("the space is not in the list", func() {
								BeforeEach(func() {
									input.Write([]byte("invalid-space\n"))
								})

								It("returns an error", func() {
									Expect(executeErr).To(MatchError(translatableerror.SpaceNotFoundError{Name: "invalid-space"}))
								})

								It("does not target the space", func() {
									Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
								})
							})

							When("the user exits the prompt early", func() {
								var fakeUI *commandfakes.FakeUI

								BeforeEach(func() {
									fakeUI = new(commandfakes.FakeUI)
									cmd.UI = fakeUI
								})

								When("the prompt returns with an EOF", func() {
									BeforeEach(func() {
										fakeUI.DisplayTextMenuReturns("", io.EOF)
									})
									It("selects no space and returns no error", func() {
										Expect(executeErr).ToNot(HaveOccurred())
										Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
									})
								})

							})

						})

					})
				})
			})
		})
	})
})
