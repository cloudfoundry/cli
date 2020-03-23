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
		fakeActor  *v7fakes.FakeActor
		fakeConfig *commandfakes.FakeConfig
		err        error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v7fakes.FakeActor)
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = APICommand{
			BaseCommand: BaseCommand{
				UI:     testUI,
				Actor:  fakeActor,
				Config: fakeConfig,
			},
		}

		fakeConfig.BinaryNameReturns("faceman")
		fakeConfig.TargetReturns("some-api-target")
	})

	JustBeforeEach(func() {
		err = cmd.Execute(nil)
	})

	When("clearing the target (the --unset flag is given)", func() {
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

	When("setting a new API endpoint (a valid API endpoint is provided)", func() {
		var CCAPI string

		BeforeEach(func() {
			CCAPI = "https://api.foo.com"
			cmd.OptionalArgs.URL = CCAPI

			fakeConfig.APIVersionReturns("100.200.300")
		})

		It("signals that it's setting the API endpoint", func() {
			Expect(testUI.Out).To(Say("Setting api endpoint to %s...", CCAPI))
		})

		When("the API endpoint does not specify a protocol/schema (does not begin with 'http')", func() {
			BeforeEach(func() {
				CCAPI = "api.foo.com"
				cmd.OptionalArgs.URL = CCAPI
			})

			It("defaults to TLS (prepends 'https://' to the endpoint)", func() {
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

		When("--skip-ssl-validation is passed", func() {
			BeforeEach(func() {
				cmd.SkipSSLValidation = true
			})

			It("skips SSL validation", func() {
				Expect(err).ToNot(HaveOccurred())

				settings := fakeActor.SetTargetArgsForCall(0)
				Expect(settings.SkipSSLValidation).To(BeTrue())

				Expect(testUI.Out).To(Say(`OK

api endpoint:   some-api-target
api version:    100.200.300`,
				))
			})
		})

		When("when the endpoint is TLS but the certificate is unverified", func() {
			BeforeEach(func() {
				fakeActor.SetTargetReturns(nil, ccerror.UnverifiedServerError{URL: CCAPI})
			})

			It("returns an error with a --skip-ssl-validation tip", func() {
				Expect(err).To(MatchError(ccerror.UnverifiedServerError{URL: CCAPI}))
				Expect(testUI.Out).ToNot(Say(`api endpoint:\s+some-api-target`))
			})
		})

		When("the endpoint specifies a non-TLS URL ('http://')", func() {
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

		When("the URL host does not exist", func() {
			var (
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

		When("the API is set but the user is not logged in", func() {
			It("outputs a 'not logged in' message", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Not logged in. Use 'faceman login' or 'faceman login --sso' to log in."))
			})
		})
	})

	When("viewing the current API target (no endpoint is provided)", func() {
		When("the API target is not set", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("")
			})

			It("informs the user that the API endpoint is not set through a tip", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("No api endpoint set. Use 'cf api' to set an endpoint"))
			})
		})

		When("the API is set, the user is logged in", func() {
			BeforeEach(func() {
				fakeConfig.APIVersionReturns("100.200.300")
				fakeConfig.CurrentUserReturns(configv3.User{
					Name: "admin",
				}, nil)
			})

			It("outputs target information", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(`api endpoint:\s+some-api-target`))
				Expect(testUI.Out).To(Say(`api version:\s+100.200.300`))
			})
		})

		When("the API is set but the user is not logged in", func() {
			It("outputs a 'not logged in' message", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Not logged in. Use 'faceman login' or 'faceman login --sso' to log in."))
			})
		})
	})
})
