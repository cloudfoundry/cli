package v6_test

import (
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
		cmd        LoginCommand
		testUI     *ui.UI
		fakeActor  *v6fakes.FakeLoginActor
		fakeConfig *commandfakes.FakeConfig
		executeErr error
		input      *Buffer
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeActor = new(v6fakes.FakeLoginActor)
		fakeConfig = new(commandfakes.FakeConfig)

		cmd = LoginCommand{
			UI:     testUI,
			Actor:  fakeActor,
			Config: fakeConfig,
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
				When("the user enters something at the prompt", func() {
					BeforeEach(func() {
						input.Write([]byte("api.boshlite.com\n"))
						cmd.APIEndpoint = ""
					})

					It("targets the API that the user inputted", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("API endpoint:"))
						Expect(testUI.Out).To(Say("api.boshlite.com\n"))
						Expect(testUI.Out).To(Say(`API endpoint:\s+api\.boshlite\.com \(API Version: 3\.4\.5\)`))

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
						Expect(testUI.Out).To(Say(`API endpoint:\s+api\.boshlite\.com \(API Version: 3\.4\.5\)`))
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

					Expect(testUI.Out).To(Say(`API endpoint:\s+api\.boshlite\.com \(API Version: 3\.4\.5\)`))
				})
			})
		})
	})
})
