package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("api Command", func() {
	var (
		cmd        APICommand
		testUI     *ui.UI
		fakeActor  *v7fakes.FakeAPIActor
		fakeConfig *commandfakes.FakeConfig
		err        error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v7fakes.FakeAPIActor)
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = APICommand{
			UI:     testUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}

		fakeConfig.BinaryNameReturns("faceman")
	})

	JustBeforeEach(func() {
		err = cmd.Execute(nil)
	})

	When("the API endpoint is not provided", func() {
		When("the API is not set", func() {
			It("displays a tip", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("No api endpoint set. Use 'cf api' to set an endpoint"))
			})
		})

		When("the API is set, the user is logged in and an org and space are targeted", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("some-api-target")
				fakeConfig.APIVersionReturns("100.200.300")
				fakeConfig.CurrentUserReturns(configv3.User{
					Name: "admin",
				}, nil)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: "some-org",
				})
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: "some-space",
				})
			})

			It("outputs target information", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(`api endpoint:\s+some-api-target`))
				Expect(testUI.Out).To(Say(`api version:\s+100.200.300`))
			})
		})

		When("passed a --unset", func() {
			BeforeEach(func() {
				cmd.Unset = true
			})

			It("clears the target", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say("Unsetting api endpoint..."))
				Expect(testUI.Out).To(Say("OK"))
				Expect(fakeActor.ClearTargetCallCount()).To(Equal(1))
			})
		})
	})

	When("a valid API endpoint is provided", func() {
		When("the API has SSL", func() {
			Context("with no protocol", func() {
				var (
					CCAPI string
				)

				BeforeEach(func() {
					CCAPI = "api.foo.com"
					cmd.OptionalArgs.URL = CCAPI

					fakeConfig.TargetReturns("some-api-target")
					fakeConfig.APIVersionReturns("100.200.300")
				})

				When("the url has verified SSL", func() {
					It("sets the target", func() {
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
						settings := fakeActor.SetTargetArgsForCall(0)
						Expect(settings.URL).To(Equal("https://" + CCAPI))
						Expect(settings.SkipSSLValidation).To(BeFalse())

						Expect(testUI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
						Expect(testUI.Out).To(Say(`OK

api endpoint:   some-api-target
api version:    100.200.300`,
						))
					})
				})

				When("the url has unverified SSL", func() {
					When("--skip-ssl-validation is passed", func() {
						BeforeEach(func() {
							cmd.SkipSSLValidation = true
						})

						It("sets the target", func() {
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
							settings := fakeActor.SetTargetArgsForCall(0)
							Expect(settings.URL).To(Equal("https://" + CCAPI))
							Expect(settings.SkipSSLValidation).To(BeTrue())

							Expect(testUI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
							Expect(testUI.Out).To(Say(`OK

api endpoint:   some-api-target
api version:    100.200.300`,
							))
						})
					})

					When("no additional flags are passed", func() {
						BeforeEach(func() {
							fakeActor.SetTargetReturns(nil, ccerror.UnverifiedServerError{URL: CCAPI})
						})

						It("returns an error with a --skip-ssl-validation tip", func() {
							Expect(err).To(MatchError(ccerror.UnverifiedServerError{URL: CCAPI}))
							Expect(testUI.Out).ToNot(Say(`api endpoint:\s+some-api-target`))
						})
					})
				})
			})
		})

		When("the API does not have SSL", func() {
			var CCAPI string

			BeforeEach(func() {
				CCAPI = "http://api.foo.com"
				cmd.OptionalArgs.URL = CCAPI
			})

			It("sets the target with a warning", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
				settings := fakeActor.SetTargetArgsForCall(0)
				Expect(settings.URL).To(Equal(CCAPI))
				Expect(settings.SkipSSLValidation).To(BeFalse())

				Expect(testUI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
				Expect(testUI.Out).To(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the API is set but the user is not logged in", func() {
			BeforeEach(func() {
				cmd.OptionalArgs.URL = "https://api.foo.com"
				fakeConfig.TargetReturns("something")
			})

			It("outputs a 'not logged in' message", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Not logged in. Use 'faceman login' or 'faceman login --sso' to log in."))
			})
		})

		When("the API is set but the user is logged in", func() {
			BeforeEach(func() {
				cmd.OptionalArgs.URL = "https://api.foo.com"
				fakeConfig.TargetReturns("something")
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			It("does not output a 'not logged in' message", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).ToNot(Say("Not logged in. Use 'faceman login' or 'faceman login --sso' to log in."))
			})
		})

		When("the API has an older version", func() {
			BeforeEach(func() {
				cmd.OptionalArgs.URL = "https://api.foo.com"
				fakeConfig.TargetReturns("something")
				fakeConfig.APIVersionReturns("1.2.3")
			})

			It("outputs a warning", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("Your CF API version .+ is no longer supported. Upgrade to a newer version of the API .+"))
			})
		})

		When("the URL host does not exist", func() {
			var (
				CCAPI      string
				requestErr ccerror.RequestError
			)

			BeforeEach(func() {
				CCAPI = "i.do.not.exist.com"
				cmd.OptionalArgs.URL = CCAPI

				requestErr = ccerror.RequestError{Err: errors.New("I am an error")}
				fakeActor.SetTargetReturns(nil, requestErr)
			})

			It("returns an APIRequestError", func() {
				Expect(err).To(MatchError(ccerror.RequestError{Err: requestErr.Err}))
			})
		})
	})
})
