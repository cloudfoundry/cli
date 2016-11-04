package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/commands/commandsfakes"
	. "code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/commands/v2/common"
	"code.cloudfoundry.org/cli/commands/v2/v2fakes"
	"code.cloudfoundry.org/cli/utils/configv3"
	"code.cloudfoundry.org/cli/utils/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("API Command", func() {
	var (
		cmd        ApiCommand
		fakeUI     *ui.UI
		fakeActor  *v2fakes.FakeAPIConfigActor
		fakeConfig *commandsfakes.FakeConfig
		err        error
	)

	BeforeEach(func() {
		out := NewBuffer()
		fakeUI = ui.NewTestUI(nil, out, out)
		fakeActor = new(v2fakes.FakeAPIConfigActor)
		fakeConfig = new(commandsfakes.FakeConfig)
		fakeConfig.ExperimentalReturns(true)

		cmd = ApiCommand{
			UI:     fakeUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
	})

	JustBeforeEach(func() {
		err = cmd.Execute(nil)
	})

	It("Displays the experimental warning message", func() {
		Expect(fakeUI.Out).To(Say(ExperimentalWarning))
	})

	Context("when the API URL is not provided", func() {

		Context("when the API is not set", func() {
			It("displays a tip", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("No api endpoint set. Use 'cf api' to set an endpoint"))
			})
		})

		Context("when the API is set", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("some-api-target")
				fakeConfig.APIVersionReturns("some-version")
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: "some-org",
				})
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: "some-space",
				})
				fakeConfig.CurrentUserReturns(configv3.User{
					Name: "admin",
				}, nil)
			})

			It("outputs the standard target information", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUI.Out).To(Say("API endpoint:\\s+some-api-target"))
				Expect(fakeUI.Out).To(Say("API version:\\s+some-version"))
				Expect(fakeUI.Out).To(Say("User:\\s+admin"))
				Expect(fakeUI.Out).To(Say("Org:\\s+some-org"))
				Expect(fakeUI.Out).To(Say("Space:\\s+some-space"))
			})
		})

		Context("when passed a --unset", func() {
			BeforeEach(func() {
				cmd.Unset = true
			})

			It("clears the target", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUI.Out).To(Say("Unsetting api endpoint..."))
				Expect(fakeUI.Out).To(Say("OK"))
				Expect(fakeActor.ClearTargetCallCount()).To(Equal(1))
			})
		})
	})

	Context("when a valid api endpoint is specified", func() {
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

						Expect(fakeUI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
						Expect(fakeUI.Out).To(Say("OK"))
						Expect(fakeUI.Out).To(Say("API endpoint:\\s+some-api-target"))
						Expect(fakeUI.Out).To(Say("API version:\\s+some-version"))
						Expect(fakeUI.Out).To(Say("User:"))
						Expect(fakeUI.Out).To(Say("Org:"))
						Expect(fakeUI.Out).To(Say("Space:"))
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

							Expect(fakeUI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
							Expect(fakeUI.Out).To(Say("OK"))
							Expect(fakeUI.Out).To(Say("API endpoint:\\s+some-api-target"))
							Expect(fakeUI.Out).To(Say("API version:\\s+some-version"))
							Expect(fakeUI.Out).To(Say("User:"))
							Expect(fakeUI.Out).To(Say("Org:"))
							Expect(fakeUI.Out).To(Say("Space:"))
						})
					})

					Context("when no additional flags are passed", func() {
						BeforeEach(func() {
							fakeActor.SetTargetReturns(nil, cloudcontroller.UnverifiedServerError{URL: CCAPI})
						})

						It("returns an error with a --skip-ssl-validation tip", func() {
							Expect(err).To(MatchError(common.InvalidSSLCertError{API: CCAPI}))
							Expect(fakeUI.Out).ToNot(Say("API endpoint:\\s+some-api-target"))
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

					Expect(fakeUI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
					Expect(fakeUI.Out).To(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
					Expect(fakeUI.Out).To(Say("OK"))
				})
			})
		})

		Context("when URL host does not exist", func() {
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

			It("sets the target with a warning", func() {
				Expect(err).To(MatchError(common.APIRequestError{Err: requestErr.Err}))
			})
		})
	})
})
