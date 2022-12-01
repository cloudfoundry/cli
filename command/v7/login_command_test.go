package v7_test

import (
	"errors"
	"fmt"
	"io"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("login Command", func() {
	var (
		binaryName        string
		cmd               LoginCommand
		testUI            *ui.UI
		fakeActor         *v7fakes.FakeActor
		fakeConfig        *commandfakes.FakeConfig
		fakeActorReloader *v7fakes.FakeActorReloader
		executeErr        error
		input             *Buffer
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v7fakes.FakeActor)
		fakeActorReloader = new(v7fakes.FakeActorReloader)

		binaryName = "some-executable"
		fakeConfig.BinaryNameReturns(binaryName)

		fakeConfig.UAAOAuthClientReturns("cf")

		cmd = LoginCommand{
			UI:            testUI,
			Actor:         fakeActor,
			Config:        fakeConfig,
			ActorReloader: fakeActorReloader,
		}
		cmd.APIEndpoint = ""

		fakeActorReloader.ReloadReturns(fakeActor, nil)
		fakeConfig.APIVersionReturns("3.99.0")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Describe("Validations", func() {
		When("the --sso and the --origin flag are used together", func() {
			BeforeEach(func() {
				cmd.SSO = true
				cmd.Origin = "some-origin"
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--sso", "--origin"},
				}))
			})
		})

		When("the --sso-passcode and the --origin flag are used together", func() {
			BeforeEach(func() {
				cmd.SSOPasscode = "some-passcode"
				cmd.Origin = "some-origin"
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--sso-passcode", "--origin"},
				}))
			})
		})

		When("the --sso and the --sso-passcode flag are used together", func() {
			BeforeEach(func() {
				cmd.SSO = true
				cmd.SSOPasscode = "some-passcode"
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--sso-passcode", "--sso"},
				}))
			})
		})

		When("the user has manually added their client credentials", func() {
			BeforeEach(func() {
				fakeConfig.UAAOAuthClientReturns("some-other-client-id")
				fakeConfig.UAAOAuthClientSecretReturns("some-secret")
			})

			It("returns an unsupported error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ManualClientCredentialsError{}))
			})
		})

		When("the current grant type is client credentials", func() {
			BeforeEach(func() {
				fakeConfig.UAAGrantTypeReturns(string(constant.GrantTypeClientCredentials))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.PasswordGrantTypeLogoutRequiredError{}))
			})
		})

		When("running against cf-on-k8s API", func() {
			BeforeEach(func() {
				fakeConfig.IsCFOnK8sReturns(true)
				fakeConfig.TargetReturns("https://foo.bar")
			})

			When("password flag is provider", func() {
				BeforeEach(func() {
					cmd.Password = "pass"
				})

				It("returns unsupported flag error", func() {
					Expect(executeErr).To(Equal(translatableerror.NotSupportedOnKubernetesArgumentError{Arg: "-p"}))
				})
			})

			When("sso flag is provider", func() {
				BeforeEach(func() {
					cmd.SSO = true
				})

				It("returns unsupported flag error", func() {
					Expect(executeErr).To(Equal(translatableerror.NotSupportedOnKubernetesArgumentError{Arg: "--sso"}))
				})
			})

			When("sso passcode flag is provider", func() {
				BeforeEach(func() {
					cmd.SSOPasscode = "sso-pass"
				})

				It("returns unsupported flag error", func() {
					Expect(executeErr).To(Equal(translatableerror.NotSupportedOnKubernetesArgumentError{Arg: "--sso-passcode"}))
				})
			})

			When("username flag is provider", func() {
				BeforeEach(func() {
					cmd.Username = "my-user"
				})

				It("returns unsupported flag error", func() {
					Expect(executeErr).To(Equal(translatableerror.NotSupportedOnKubernetesArgumentError{Arg: "-u"}))
				})
			})

			When("origin flag is provider", func() {
				BeforeEach(func() {
					cmd.Origin = "my-origin"
				})

				It("returns unsupported flag error", func() {
					Expect(executeErr).To(Equal(translatableerror.NotSupportedOnKubernetesArgumentError{Arg: "--origin"}))
				})
			})
		})
	})

	Describe("API Endpoint", func() {
		When("user provides the api endpoint using the -a flag", func() {
			BeforeEach(func() {
				fakeActor.SetTargetReturns(v7action.Warnings{"some-warning-1", "some-warning-2"}, nil)
				cmd.APIEndpoint = "api.example.com"
			})

			It("targets the provided api endpoint and prints all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.SetTargetCallCount()).To(Equal(1))

				actualSettings := fakeActor.SetTargetArgsForCall(0)
				Expect(actualSettings.URL).To(Equal("https://api.example.com"))
				Expect(actualSettings.SkipSSLValidation).To(Equal(false))

				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			When("the user specifies --skip-ssl-validation", func() {
				BeforeEach(func() {
					cmd.SkipSSLValidation = true
				})

				It("targets the provided api endpoint", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					actualSettings := fakeActor.SetTargetArgsForCall(0)
					Expect(actualSettings.URL).To(Equal("https://api.example.com"))
					Expect(actualSettings.SkipSSLValidation).To(Equal(true))
				})
			})

			When("targeting the API fails", func() {
				BeforeEach(func() {
					fakeActor.SetTargetReturns(
						v7action.Warnings{"some-warning-1", "some-warning-2"},
						errors.New("some error"))
				})

				It("errors and prints all warnings", func() {
					Expect(executeErr).To(MatchError("some error"))
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})
		})

		When("user does not provide the api endpoint using the -a flag", func() {
			When("config has API endpoint already set", func() {
				BeforeEach(func() {
					fakeConfig.TargetReturns("api.fake.com")
				})

				It("uses the API endpoint from the config", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.SetTargetCallCount()).To(Equal(1))

					actualSettings := fakeActor.SetTargetArgsForCall(0)
					Expect(actualSettings.URL).To(Equal("https://api.fake.com"))
				})

				When("the API version is older than the minimum supported API version for the v7 CLI", func() {
					BeforeEach(func() {
						fakeConfig.APIVersionReturns("3.83.0")
					})
					It("warns that the user is targeting an unsupported API version and that things may not work correctly", func() {
						Expect(testUI.Err).To(Say("Warning: Your targeted API's version \\(3.83.0\\) is less than the minimum supported API version \\(3.99.0\\). Some commands may not function correctly."))
					})
				})

				When("the API version is empty", func() {
					BeforeEach(func() {
						fakeConfig.APIVersionReturns("")
					})
					It("prints a warning message", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Err).To(Say("Warning: unable to determine whether targeted API's version meets minimum supported."))
					})
				})

				It("should NOT warn that the user is targeting an unsupported API version", func() {
					Expect(testUI.Err).ToNot(Say("is less than the minimum supported API version"))
				})

				When("the config has SkipSSLValidation false and the --skip-ssl-validation flag is passed", func() {
					BeforeEach(func() {
						fakeConfig.SkipSSLValidationReturns(false)
						cmd.SkipSSLValidation = true
					})

					It("sets the target with SkipSSLValidation is true", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
						targetSettings := fakeActor.SetTargetArgsForCall(0)
						Expect(targetSettings.SkipSSLValidation).To(BeTrue())
					})
				})
			})

			When("config does not have an API endpoint set and the user enters the endpoint at the prompt", func() {
				BeforeEach(func() {
					cmd.APIEndpoint = ""
					_, err := input.Write([]byte("api.example.com\n"))
					Expect(err).ToNot(HaveOccurred())
					fakeConfig.TargetReturnsOnCall(0, "")
					fakeConfig.TargetReturnsOnCall(1, "https://api.example.com")
				})

				It("targets the API that the user inputted", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
					actualSettings := fakeActor.SetTargetArgsForCall(0)
					Expect(actualSettings.URL).To(Equal("https://api.example.com"))
					Expect(actualSettings.SkipSSLValidation).To(Equal(false))
					Expect(fakeConfig.TargetCallCount()).To(Equal(2))
				})

				When("the user specifies --skip-ssl-validation", func() {
					BeforeEach(func() {
						cmd.SkipSSLValidation = true
					})

					It("targets the API that the user inputted", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						actualSettings := fakeActor.SetTargetArgsForCall(0)
						Expect(actualSettings.SkipSSLValidation).To(Equal(true))
					})
				})
			})
		})

		When("the endpoint has trailing slashes", func() {
			BeforeEach(func() {
				cmd.APIEndpoint = "api.example.com////"
				fakeConfig.TargetReturns("https://api.example.com///")
			})

			It("strips the backslashes before using the endpoint", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
				actualSettings := fakeActor.SetTargetArgsForCall(0)
				Expect(actualSettings.URL).To(Equal("https://api.example.com"))
			})
		})

		When("targeting the API fails due to an invalid certificate", func() {
			BeforeEach(func() {
				cmd.APIEndpoint = "api.example.com"
				fakeActor.SetTargetReturns(nil, ccerror.UnverifiedServerError{URL: "https://api.example.com"})
			})

			It("returns an error mentioning the login command", func() {
				Expect(executeErr).To(MatchError(
					translatableerror.InvalidSSLCertError{URL: "https://api.example.com", SuggestedCommand: "login"}))
			})
		})
	})

	Describe("username and password", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("https://some.random.endpoint")
		})

		When("the current grant type is password", func() {
			BeforeEach(func() {
				fakeConfig.UAAGrantTypeReturns(string(constant.GrantTypePassword))
			})

			It("fetches prompts from the UAA", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetLoginPromptsCallCount()).To(Equal(1))
			})

			When("one of the prompts has a username key and is text type", func() {
				BeforeEach(func() {
					fakeActor.GetLoginPromptsReturns(map[string]coreconfig.AuthPrompt{
						"username": {
							DisplayName: "Username",
							Type:        coreconfig.AuthPromptTypeText,
						},
					}, nil)
				})

				When("the username flag is set", func() {
					BeforeEach(func() {
						cmd.Username = "potatoface"
					})

					It("uses the provided value for the username", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
						credentials, _, _ := fakeActor.AuthenticateArgsForCall(0)
						Expect(credentials["username"]).To(Equal("potatoface"))
					})

					When("the --origin flag is set", func() {
						BeforeEach(func() {
							cmd.Origin = "picklebike"
						})

						It("authenticates with the specified origin", func() {
							Expect(executeErr).NotTo(HaveOccurred())
							Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
							credentials, origin, _ := fakeActor.AuthenticateArgsForCall(0)
							Expect(credentials["username"]).To(Equal("potatoface"))
							Expect(origin).To(Equal("picklebike"))
						})
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
					}, nil)
				})

				When("the password flag is set", func() {
					BeforeEach(func() {
						cmd.Password = "noprompto"
					})

					It("uses the provided value", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
						credentials, _, _ := fakeActor.AuthenticateArgsForCall(0)
						Expect(credentials["password"]).To(Equal("noprompto"))
					})

					When("the password is incorrect", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("other-password\n"))
							Expect(err).ToNot(HaveOccurred())
							fakeActor.AuthenticateReturnsOnCall(0, errors.New("bad creds"))
							fakeActor.AuthenticateReturnsOnCall(1, nil)
						})

						It("does not reuse the flag value for subsequent attempts", func() {
							credentials, _, _ := fakeActor.AuthenticateArgsForCall(1)
							Expect(credentials["password"]).To(Equal("other-password"))
						})
					})

					When("there have been too many failed login attempts", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("other-password\n"))
							Expect(err).ToNot(HaveOccurred())
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
					}, nil)
				})

				It("does not prompt for the passcode, only the password", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).NotTo(Say("gimme your passcode"))
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
					}, nil)
				})

				When("all authentication information is coming from prompts, not flags", func() {
					BeforeEach(func() {
						_, err := input.Write([]byte("faker\nsomeaccount\nsomepassword\ngarbage\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("displays text prompts, starting with username, then password prompts, starting with password", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("\n"))
						Expect(testUI.Out).To(Say("Username:"))
						Expect(testUI.Out).To(Say("faker"))

						Expect(testUI.Out).To(Say("\n"))
						Expect(testUI.Out).To(Say("Account Number:"))
						Expect(testUI.Out).To(Say("someaccount"))

						Expect(testUI.Out).To(Say("\n"))
						Expect(testUI.Out).To(Say("Your Password:"))
						Expect(testUI.Out).NotTo(Say("somepassword"))

						Expect(testUI.Out).To(Say("\n"))
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
							UI:            fakeUI,
							Actor:         fakeActor,
							Config:        fakeConfig,
							ActorReloader: fakeActorReloader,
						}
					})

					It("stops prompting after the first prompt and errors", func() {
						Expect(fakeUI.DisplayTextPromptCallCount()).To(Equal(1))
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
							UI:            fakeUI,
							Actor:         fakeActor,
							Config:        fakeConfig,
							ActorReloader: fakeActorReloader,
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
							UI:            fakeUI,
							Actor:         fakeActor,
							Config:        fakeConfig,
							ActorReloader: fakeActorReloader,
						}
					})

					It("stops prompting after the first prompt and errors", func() {
						Expect(fakeUI.DisplayPasswordPromptCallCount()).To(Equal(1))
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
							UI:            fakeUI,
							Actor:         fakeActor,
							Config:        fakeConfig,
							ActorReloader: fakeActorReloader,
						}
					})

					It("stops prompting after the second prompt and errors", func() {
						Expect(executeErr).To(MatchError("Unable to authenticate."))
					})
				})

				When("authenticating succeeds", func() {
					BeforeEach(func() {
						fakeActor.GetCurrentUserReturns(configv3.User{Name: "potatoface"}, nil)
						_, err := input.Write([]byte("faker\nsomeaccount\nsomepassword\ngarbage\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("displays OK and a status summary", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
						Expect(testUI.Out).To(Say(`user:\s+potatoface`))

						Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
					})
				})

				When("authenticating fails", func() {
					BeforeEach(func() {
						fakeActor.AuthenticateReturns(errors.New("something died"))
						_, err := input.Write([]byte("faker\nsomeaccount\nsomepassword\ngarbage\nfaker\nsomeaccount\nsomepassword\ngarbage\nfaker\nsomeaccount\nsomepassword\ngarbage\n"))
						Expect(err).ToNot(HaveOccurred())
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

						Expect(executeErr).To(MatchError("Unable to authenticate."))
						Expect(fakeActor.AuthenticateCallCount()).To(Equal(3))
					})

					It("displays a status summary", func() {
						Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, cmd.APIEndpoint))
						Expect(testUI.Out).To(Say(`Not logged in. Use '%s login' or '%s login --sso' to log in.`, cmd.Config.BinaryName(), cmd.Config.BinaryName()))
					})
				})

				When("authenticating fails with a bad credentials error", func() {
					BeforeEach(func() {
						fakeActor.AuthenticateReturns(uaa.UnauthorizedError{Message: "Bad credentials"})
						_, err := input.Write([]byte("faker\nsomeaccount\nsomepassword\ngarbage\nfaker\nsomeaccount\nsomepassword\ngarbage\nfaker\nsomeaccount\nsomepassword\ngarbage\n"))
						Expect(err).ToNot(HaveOccurred())
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
		})
	})

	Describe("SSO Passcode", func() {
		fakeAPI := "whatever.com"
		BeforeEach(func() {
			fakeConfig.TargetReturns(fakeAPI)

			_, err := input.Write([]byte("some-passcode\n"))
			Expect(err).ToNot(HaveOccurred())
			fakeActor.GetLoginPromptsReturns(map[string]coreconfig.AuthPrompt{
				"passcode": {
					DisplayName: "some-sso-prompt",
					Type:        coreconfig.AuthPromptTypePassword,
				},
			}, nil)

			fakeActor.GetCurrentUserReturns(configv3.User{Name: "potatoface"}, nil)
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
				Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, fakeAPI))
				Expect(testUI.Out).To(Say(`user:\s+potatoface`))

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
						UI:            fakeUI,
						Actor:         fakeActor,
						Config:        fakeConfig,
						ActorReloader: fakeActorReloader,
						SSO:           true,
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
				Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, fakeAPI))
				Expect(testUI.Out).To(Say(`user:\s+potatoface`))
			})

			When("an incorrect passcode is inputted", func() {
				BeforeEach(func() {
					cmd.SSOPasscode = "some-garbage"
					fakeActor.AuthenticateReturns(uaa.UnauthorizedError{
						Message: "Bad credentials",
					})
					fakeActor.GetCurrentUserReturns(configv3.User{}, nil)
					_, err := input.Write([]byte("some-passcode\n"))
					Expect(err).ToNot(HaveOccurred())
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
					Expect(testUI.Out).To(Say(`API endpoint:\s+%s`, fakeAPI))
					Expect(testUI.Out).To(Say(`Not logged in. Use '%s login' or '%s login --sso' to log in.`, cmd.Config.BinaryName(), cmd.Config.BinaryName()))
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

	Describe("Config", func() {
		When("a user has successfully authenticated", func() {
			BeforeEach(func() {
				cmd.APIEndpoint = "example.com"
				cmd.Username = "some-user"
				cmd.Password = "some-password"
				fakeConfig.APIVersionReturns("3.4.5")
				fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			})

			It("writes to the config", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeConfig.WriteConfigCallCount()).To(Equal(1))
			})

			When("GetOrganizations fails", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationsReturns(nil, nil, errors.New("Org Failure"))
				})
				It("writes to the config", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(fakeConfig.WriteConfigCallCount()).To(Equal(1))
				})
			})

			When("WriteConfig returns an error", func() {
				BeforeEach(func() {
					fakeConfig.WriteConfigReturns(errors.New("Config Failure"))
				})
				It("throws that error", func() {
					Expect(executeErr).To(MatchError("Error writing config: Config Failure"))
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
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
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
						resources.Organization{Name: "some-org", GUID: "some-guid"},
						v7action.Warnings{"some-warning-1", "some-warning-2"},
						nil)
					fakeConfig.TargetedOrganizationNameReturns("some-org")
					fakeConfig.TargetReturns("https://example.com")
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
					Expect(testUI.Out).To(Say(`API endpoint:\s+https://example.com`))
					Expect(testUI.Out).To(Say(`API version:\s+3.4.5`))
					Expect(testUI.Out).To(Say("user:           some-user"))
					Expect(testUI.Out).To(Say("org:            some-org"))
				})
			})

			When("fetching the organization fails", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{},
						v7action.Warnings{"some-warning-1", "some-warning-2"},
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

		When("-o was not passed, -s was passed", func() {
			BeforeEach(func() {
				cmd.APIEndpoint = "example.com"
				cmd.Username = "some-user"
				cmd.Password = "some-password"
				cmd.Space = "some-space"
				fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
				fakeConfig.TargetReturns("https://example.com")
				fakeActor.GetOrganizationsReturns(
					[]resources.Organization{},
					v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
					nil,
				)
				fakeActor.GetSpaceByNameAndOrganizationCalls(func(spaceName string, orgGUID string) (resources.Space, v7action.Warnings, error) {
					if orgGUID != "some-org-guid1" {
						return resources.Space{Name: spaceName}, v7action.Warnings{}, nil
					}
					return resources.Space{}, v7action.Warnings{}, actionerror.SpaceNotFoundError{}
				})
			})

			When("no org valid org exists", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationsReturns(
						[]resources.Organization{{
							GUID: "some-org-guid",
							Name: "some-org-name",
						}},
						v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
						nil,
					)
				})

				It("does not prompt the user to select an org", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).ToNot(Say("Select an org:"))
					Expect(testUI.Out).ToNot(Say(`Org \(enter to skip\):`))
				})

				It("displays how to target an org and space", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say(`API endpoint:\s+https://example.com`))
					Expect(testUI.Out).To(Say(`API version:\s+3.4.5`))
					Expect(testUI.Out).To(Say(`user:\s+some-user`))
					Expect(testUI.Out).To(Say("No org or space targeted, use '%s target -o ORG -s SPACE'", binaryName))
				})
			})

			When("only one valid org exists", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationsReturns(
						[]resources.Organization{{
							GUID: "some-org-guid1",
							Name: "some-org-name1",
						}, {
							GUID: "some-org-guid2",
							Name: "some-org-name2",
						}},
						v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
						nil,
					)
				})

				It("targets that org", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
					orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid2"))
					Expect(orgName).To(Equal("some-org-name2"))
				})
			})

			When("more than one valid org exists", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationsReturns(
						[]resources.Organization{
							{
								GUID: "some-org-guid3",
								Name: "1234",
							},
							{
								GUID: "some-org-guid1",
								Name: "some-org-name1",
							},
							{
								GUID: "some-org-guid2",
								Name: "some-org-name2",
							},
						},
						v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
						nil,
					)
				})

				It("prompts the user to select an org from the filtered selection", func() {
					Expect(testUI.Out).To(Say("Select an org:"))
					Expect(testUI.Out).To(Say("1. 1234"))
					Expect(testUI.Out).To(Say("2. some-org-name2"))
					Expect(testUI.Out).To(Say("\n\n"))
					Expect(testUI.Out).To(Say(`Org \(enter to skip\):`))
					Expect(executeErr).ToNot(HaveOccurred())
				})
			})

			When("filtering the orgs errors", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationsReturns(
						[]resources.Organization{{
							GUID: "some-org-guid",
							Name: "some-org-name",
						}},
						v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
						nil,
					)
					fakeActor.GetSpaceByNameAndOrganizationReturns(resources.Space{}, v7action.Warnings{}, errors.New("oh noooooooo"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(errors.New("oh noooooooo")))
				})
			})
		})

		When("-o and -s were both not passed", func() {
			BeforeEach(func() {
				cmd.APIEndpoint = "example.com"
				cmd.Username = "some-user"
				cmd.Password = "some-password"
				fakeActor.GetOrganizationsReturns(
					[]resources.Organization{},
					v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
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
					fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
					fakeConfig.TargetReturns("https://example.com")
				})

				When("no org exists", func() {
					It("does not prompt the user to select an org", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).ToNot(Say("Select an org:"))
						Expect(testUI.Out).ToNot(Say(`Org \(enter to skip\):`))
					})

					It("displays how to target an org and space", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say(`API endpoint:\s+https://example.com`))
						Expect(testUI.Out).To(Say(`API version:\s+3.4.5`))
						Expect(testUI.Out).To(Say(`user:\s+some-user`))
						Expect(testUI.Out).To(Say("No org or space targeted, use '%s target -o ORG -s SPACE'", binaryName))
					})
				})

				When("only one org exists", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationsReturns(
							[]resources.Organization{{
								GUID: "some-org-guid",
								Name: "some-org-name",
							}},
							v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
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
							[]resources.Organization{
								{
									GUID: "some-org-guid3",
									Name: "1234",
								},
								{
									GUID: "some-org-guid1",
									Name: "some-org-name1",
								},
								{
									GUID: "some-org-guid2",
									Name: "some-org-name2",
								},
							},
							v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
							nil,
						)
					})

					When("the user selects an org by list position", func() {
						When("the position is valid", func() {
							BeforeEach(func() {
								fakeConfig.TargetedOrganizationReturns(configv3.Organization{
									GUID: "targeted-org-guid1",
								})
								fakeConfig.TargetedOrganizationNameReturns("targeted-org-name")
								_, err := input.Write([]byte("2\n"))
								Expect(err).ToNot(HaveOccurred())
							})

							It("prompts the user to select an org", func() {
								Expect(testUI.Out).To(Say("Select an org:"))
								Expect(testUI.Out).To(Say("1. 1234"))
								Expect(testUI.Out).To(Say("2. some-org-name1"))
								Expect(testUI.Out).To(Say("3. some-org-name2"))
								Expect(testUI.Out).To(Say("\n\n"))
								Expect(testUI.Out).To(Say(`Org \(enter to skip\):`))
								Expect(executeErr).ToNot(HaveOccurred())
							})

							It("targets that org", func() {
								Expect(fakeConfig.SetOrganizationInformationCallCount()).To(Equal(1))
								orgGUID, orgName := fakeConfig.SetOrganizationInformationArgsForCall(0)
								Expect(orgGUID).To(Equal("some-org-guid1"))
								Expect(orgName).To(Equal("some-org-name1"))
							})

							It("outputs targeted org", func() {
								Expect(testUI.Out).To(Say("Targeted org targeted-org-name"))
							})
						})

						When("the position is invalid", func() {
							BeforeEach(func() {
								_, err := input.Write([]byte("4\n"))
								Expect(err).ToNot(HaveOccurred())
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
								_, err := input.Write([]byte("some-org-name2\n"))
								Expect(err).NotTo(HaveOccurred())
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
								_, err := input.Write([]byte("invalid-org\n"))
								Expect(err).NotTo(HaveOccurred())
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
						orgs := make([]resources.Organization, 51)
						for i := range orgs {
							orgs[i].Name = fmt.Sprintf("org%d", i+1)
							orgs[i].GUID = fmt.Sprintf("org-guid%d", i+1)
						}

						fakeActor.GetOrganizationsReturns(
							orgs,
							v7action.Warnings{"some-org-warning-1", "some-org-warning-2"},
							nil,
						)
					})

					When("the user selects an org by name", func() {
						When("the list contains that org", func() {
							BeforeEach(func() {
								_, err := input.Write([]byte("org37\n"))
								Expect(err).NotTo(HaveOccurred())
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
								_, err := input.Write([]byte("invalid-org\n"))
								Expect(err).NotTo(HaveOccurred())
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
						[]resources.Organization{},
						v7action.Warnings{"some-warning-1", "some-warning-2"},
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
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
		})

		When("an org has been successfully targeted", func() {
			BeforeEach(func() {
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					GUID: "targeted-org-guid",
					Name: "targeted-org-name",
				},
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
							resources.Space{
								Name: "some-space",
								GUID: "some-space-guid",
							},
							v7action.Warnings{"some-warning-1", "some-warning-2"},
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

						It("displays that the space has been targeted", func() {
							Expect(testUI.Out).To(Say(`space:\s+some-space`))
						})
					})
				})

				When("the specified space does not exist or does not belong to the targeted org", func() {
					BeforeEach(func() {
						fakeActor.GetSpaceByNameAndOrganizationReturns(
							resources.Space{},
							v7action.Warnings{"some-warning-1", "some-warning-2"},
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
						Expect(testUI.Out).To(Say(`space:\s+No space targeted, use 'some-executable target -s SPACE'`))
					})
				})
			})

			When("-s was not passed", func() {
				When("fetching the spaces for an organization succeeds", func() {
					When("no space exists", func() {
						BeforeEach(func() {
							fakeActor.GetOrganizationSpacesReturns(
								[]resources.Space{},
								v7action.Warnings{},
								nil,
							)
							fakeConfig.TargetReturns("https://example.com")
						})
						It("does not prompt the user to select a space", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).ToNot(Say("Select a space:"))
							Expect(testUI.Out).ToNot(Say(`Space \(enter to skip\):`))
						})

						It("displays how to target a space", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(testUI.Out).To(Say(`API endpoint:\s+https://example.com`))
							Expect(testUI.Out).To(Say(`API version:\s+3.4.5`))
							Expect(testUI.Out).To(Say(`user:\s+some-user`))
							Expect(testUI.Out).To(Say("No space targeted, use '%s target -s SPACE'", binaryName))
						})
					})

					When("only one space is available", func() {
						BeforeEach(func() {
							spaces := []resources.Space{
								{
									GUID: "some-space-guid",
									Name: "some-space-name",
								},
							}

							fakeActor.GetOrganizationSpacesReturns(
								spaces,
								v7action.Warnings{},
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

							Expect(testUI.Out).To(Say(`Targeted space some-space-name\.`))
							Expect(testUI.Out).To(Say(`space:\s+some-space-name`))
							Expect(testUI.Out).NotTo(Say(`space:\s+No space targeted, use 'some-executable target -s SPACE`))
						})
					})

					When("more than one space is available", func() {
						BeforeEach(func() {
							spaces := []resources.Space{
								{
									GUID: "some-space-guid",
									Name: "some-space-name",
								},
								{
									GUID: "some-space-guid1",
									Name: "some-space-name1",
								},
								{
									GUID: "some-space-guid2",
									Name: "some-space-name2",
								},
								{
									GUID: "some-space-guid3",
									Name: "3",
								},
								{
									GUID: "some-space-guid3",
									Name: "100",
								},
							}

							fakeActor.GetOrganizationSpacesReturns(
								spaces,
								v7action.Warnings{},
								nil,
							)
						})

						It("displays a numbered list of spaces", func() {
							Expect(testUI.Out).To(Say("Select a space:"))
							Expect(testUI.Out).To(Say("1. some-space-name"))
							Expect(testUI.Out).To(Say("2. some-space-name1"))
							Expect(testUI.Out).To(Say("3. some-space-name2"))
							Expect(testUI.Out).To(Say("4. 3"))
							Expect(testUI.Out).To(Say("5. 100"))
							Expect(testUI.Out).To(Say("\n\n"))
							Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
						})

						When("the user selects a space by list position", func() {
							When("the position is valid", func() {
								BeforeEach(func() {
									_, err := input.Write([]byte("2\n"))
									Expect(err).NotTo(HaveOccurred())
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
									_, err := input.Write([]byte("-1\n"))
									Expect(err).NotTo(HaveOccurred())
								})

								It("reprompts the user", func() {
									Expect(testUI.Out).To(Say("Select a space:"))
									Expect(testUI.Out).To(Say("1. some-space-name"))
									Expect(testUI.Out).To(Say("2. some-space-name1"))
									Expect(testUI.Out).To(Say("3. some-space-name2"))
									Expect(testUI.Out).To(Say("4. 3"))
									Expect(testUI.Out).To(Say("5. 100"))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
									Expect(testUI.Out).To(Say("Select a space:"))
									Expect(testUI.Out).To(Say("1. some-space-name"))
									Expect(testUI.Out).To(Say("2. some-space-name1"))
									Expect(testUI.Out).To(Say("3. some-space-name2"))
									Expect(testUI.Out).To(Say("4. 3"))
									Expect(testUI.Out).To(Say("5. 100"))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
								})
							})
						})

						When("the user selects a space by name", func() {
							When("the list contains that space", func() {
								BeforeEach(func() {
									_, err := input.Write([]byte("some-space-name2\n"))
									Expect(err).NotTo(HaveOccurred())
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
									_, err := input.Write([]byte("invalid-space\n"))
									Expect(err).NotTo(HaveOccurred())
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

						When("the user enters text which is both a space name and a digit", func() {
							When("the entry is a valid position", func() {
								BeforeEach(func() {
									_, err := input.Write([]byte("3\n"))
									Expect(err).NotTo(HaveOccurred())
								})

								It("targets the space at the index specified", func() {
									Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))
									guid, name, allowSSH := fakeConfig.SetSpaceInformationArgsForCall(0)
									Expect(guid).To(Equal("some-space-guid2"))
									Expect(name).To(Equal("some-space-name2"))
									Expect(allowSSH).To(BeTrue())
									Expect(executeErr).NotTo(HaveOccurred())
								})
							})

							When("the entry is an invalid position", func() {
								BeforeEach(func() {
									_, err := input.Write([]byte("100\n"))
									Expect(err).NotTo(HaveOccurred())
								})

								It("reprompts the user", func() {
									Expect(testUI.Out).To(Say("Select a space:"))
									Expect(testUI.Out).To(Say("1. some-space-name"))
									Expect(testUI.Out).To(Say("2. some-space-name1"))
									Expect(testUI.Out).To(Say("3. some-space-name2"))
									Expect(testUI.Out).To(Say("4. 3"))
									Expect(testUI.Out).To(Say("5. 100"))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
									Expect(testUI.Out).To(Say("1. some-space-name"))
									Expect(testUI.Out).To(Say("2. some-space-name1"))
									Expect(testUI.Out).To(Say("3. some-space-name2"))
									Expect(testUI.Out).To(Say("4. 3"))
									Expect(testUI.Out).To(Say("5. 100"))
									Expect(testUI.Out).To(Say("\n\n"))
									Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
								})
							})
						})
					})

					When("more than 50 spaces exist", func() {
						BeforeEach(func() {
							spaces := make([]resources.Space, 51)
							for i := range spaces {
								spaces[i].Name = fmt.Sprintf("space-%d", i+1)
								spaces[i].GUID = fmt.Sprintf("space-guid-%d", i+1)
							}

							fakeActor.GetOrganizationSpacesReturns(
								spaces,
								v7action.Warnings{},
								nil,
							)
						})

						It("prompts the user to select an space", func() {
							Expect(testUI.Out).To(Say("There are too many options to display; please type in the name."))
							Expect(testUI.Out).To(Say("\n\n"))
							Expect(testUI.Out).To(Say(`Space \(enter to skip\):`))
						})

						When("the user selects an space by name", func() {
							When("the list contains that space", func() {
								BeforeEach(func() {
									_, err := input.Write([]byte("space-37\n"))
									Expect(err).NotTo(HaveOccurred())
								})

								It("targets that space", func() {
									Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(1))
									spaceGUID, spaceName, allowSSH := fakeConfig.SetSpaceInformationArgsForCall(0)
									Expect(spaceGUID).To(Equal("space-guid-37"))
									Expect(spaceName).To(Equal("space-37"))
									Expect(allowSSH).To(BeTrue())
								})
							})

							When("the name is a valid list position, but it does not match a space name", func() {
								BeforeEach(func() {
									_, err := input.Write([]byte("31\n"))
									Expect(err).NotTo(HaveOccurred())
								})

								It("returns an error", func() {
									Expect(executeErr).To(MatchError(translatableerror.SpaceNotFoundError{Name: "31"}))
								})

								It("does not target the space", func() {
									Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
								})
							})

							When("the space is not in the list", func() {
								BeforeEach(func() {
									_, err := input.Write([]byte("invalid-space\n"))
									Expect(err).NotTo(HaveOccurred())
								})

								It("returns an error", func() {
									Expect(executeErr).To(MatchError(translatableerror.SpaceNotFoundError{Name: "invalid-space"}))
								})

								It("does not target the space", func() {
									Expect(fakeConfig.SetSpaceInformationCallCount()).To(Equal(0))
								})
							})
						})
					})
				})

				When("fetching the spaces for an organization fails", func() {
					BeforeEach(func() {
						fakeActor.GetOrganizationSpacesReturns(
							[]resources.Space{},
							v7action.Warnings{"some-warning-1", "some-warning-2"},
							errors.New("fetching spaces failed"),
						)
					})

					It("returns an error", func() {
						Expect(executeErr).To(MatchError("fetching spaces failed"))
					})

					It("returns all warnings", func() {
						Expect(testUI.Err).To(Say("some-warning-1"))
						Expect(testUI.Err).To(Say("some-warning-2"))
					})
				})
			})
		})
	})
})
