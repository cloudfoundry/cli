package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("API Command", func() {
	var (
		cmd        ApiCommand
		testUI     *ui.UI
		fakeActor  *v2fakes.FakeAPIConfigActor
		fakeConfig *commandfakes.FakeConfig
		err        error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v2fakes.FakeAPIConfigActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("faceman")

		cmd = ApiCommand{
			UI:     testUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
	})

	JustBeforeEach(func() {
		err = cmd.Execute(nil)
	})

	Context("when the API endpoint is not provided", func() {
		Context("when the API is not set", func() {
			It("displays a tip", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("No api endpoint set. Use 'cf api' to set an endpoint"))
			})
		})

		Context("when the API is set, the user is logged in and an org and space are targeted", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("some-api-target")
				fakeConfig.APIVersionReturns("some-version")
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
				Expect(testUI.Out).To(Say("API endpoint:\\s+some-api-target"))
				Expect(testUI.Out).To(Say("API version:\\s+some-version"))
			})
		})

		Context("when passed a --unset", func() {
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

	Context("when a valid API endpoint is provided", func() {
		Context("when the API has SSL", func() {
			Context("with no protocol", func() {
				var (
					CCAPI string
				)

				BeforeEach(func() {
					CCAPI = "api.foo.com"
					cmd.OptionalArgs.URL = CCAPI

					fakeConfig.TargetReturns("some-api-target")
					fakeConfig.APIVersionReturns("some-version")
				})

				Context("when the url has verified SSL", func() {
					It("sets the target", func() {
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeActor.SetTargetCallCount()).To(Equal(1))
						settings := fakeActor.SetTargetArgsForCall(0)
						Expect(settings.URL).To(Equal("https://" + CCAPI))
						Expect(settings.SkipSSLValidation).To(BeFalse())

						Expect(testUI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
						Expect(testUI.Out).To(Say(`OK

API endpoint:   some-api-target
API version:    some-version`,
						))
					})
				})

				Context("when the url has unverified SSL", func() {
					Context("when --skip-ssl-validation is passed", func() {
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

API endpoint:   some-api-target
API version:    some-version`,
							))
						})
					})

					Context("when no additional flags are passed", func() {
						BeforeEach(func() {
							fakeActor.SetTargetReturns(nil, cloudcontroller.UnverifiedServerError{URL: CCAPI})
						})

						It("returns an error with a --skip-ssl-validation tip", func() {
							Expect(err).To(MatchError(command.InvalidSSLCertError{API: CCAPI}))
							Expect(testUI.Out).ToNot(Say("API endpoint:\\s+some-api-target"))
						})
					})
				})
			})
		})

		Context("when the API does not have SSL", func() {
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

		Context("when the API is set but the user is not logged in", func() {
			BeforeEach(func() {
				cmd.OptionalArgs.URL = "https://api.foo.com"
				fakeConfig.TargetReturns("something")
			})

			It("outputs a 'not logged in' message", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Not logged in. Use 'faceman login' to log in."))
			})
		})

		Context("when the API is set but the user is logged in", func() {
			BeforeEach(func() {
				cmd.OptionalArgs.URL = "https://api.foo.com"
				fakeConfig.TargetReturns("something")
				fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			})

			It("does not output a 'not logged in' message", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).ToNot(Say("Not logged in. Use 'faceman login' to log in."))
			})
		})

		Context("when the URL host does not exist", func() {
			var (
				CCAPI      string
				requestErr cloudcontroller.RequestError
			)

			BeforeEach(func() {
				CCAPI = "i.do.not.exist.com"
				cmd.OptionalArgs.URL = CCAPI

				requestErr = cloudcontroller.RequestError{Err: errors.New("I am an error")}
				fakeActor.SetTargetReturns(nil, requestErr)
			})

			It("returns an APIRequestError", func() {
				Expect(err).To(MatchError(command.APIRequestError{Err: requestErr.Err}))
			})
		})
	})
})
