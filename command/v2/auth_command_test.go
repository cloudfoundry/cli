package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("auth Command", func() {
	var (
		cmd        AuthCommand
		testUI     *ui.UI
		fakeActor  *v2fakes.FakeAuthActor
		fakeConfig *commandfakes.FakeConfig
		binaryName string
		err        error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeActor = new(v2fakes.FakeAuthActor)
		fakeConfig = new(commandfakes.FakeConfig)
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = AuthCommand{
			UI:     testUI,
			Config: fakeConfig,
			Actor:  fakeActor,
		}
	})

	JustBeforeEach(func() {
		err = cmd.Execute(nil)
	})

	Context("when there are no errors", func() {
		var (
			testUsername string
			testPassword string
		)

		BeforeEach(func() {
			testUsername = helpers.NewUsername()
			testPassword = helpers.NewPassword()
			cmd.RequiredArgs.Username = testUsername
			cmd.RequiredArgs.Password = testPassword

			fakeConfig.TargetReturns("some-api-target")

			fakeActor.AuthenticateReturns(nil)
		})

		It("outputs API target information and clears the targeted org and space", func() {
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("API endpoint: %s", fakeConfig.Target()))
			Expect(testUI.Out).To(Say("Authenticating\\.\\.\\."))
			Expect(testUI.Out).To(Say("OK"))
			Expect(testUI.Out).To(Say("Use '%s target' to view or set your target org and space", binaryName))

			Expect(fakeActor.AuthenticateCallCount()).To(Equal(1))
			config, username, password := fakeActor.AuthenticateArgsForCall(0)
			Expect(config).To(Equal(fakeConfig))
			Expect(username).To(Equal(testUsername))
			Expect(password).To(Equal(testPassword))
		})
	})

	Context("when there is an auth error", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.Username = "foo"
			cmd.RequiredArgs.Password = "bar"

			fakeConfig.TargetReturns("some-api-target")

			fakeActor.AuthenticateReturns(uaa.BadCredentialsError{Message: "some message"})
		})

		It("returns a BadCredentialsError", func() {
			Expect(err).To(MatchError(uaa.BadCredentialsError{Message: "some message"}))
		})
	})

	Context("when there is a non-auth error", func() {
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
			fakeConfig.APIVersionReturns(apiVersion)
			minCLIVersion = "1.0.0"
			fakeConfig.MinCLIVersionReturns(minCLIVersion)
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

		Context("when the CLI version is invalid", func() {
			BeforeEach(func() {
				fakeConfig.BinaryVersionReturns("&#%")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("No Major.Minor.Patch elements found"))
			})
		})
	})
})
