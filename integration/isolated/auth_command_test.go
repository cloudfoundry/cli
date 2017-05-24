package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("auth command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("auth", "--help")
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session.Out).Should(Say("auth - Authenticate user non-interactively\n\n"))

			Eventually(session.Out).Should(Say("USAGE:"))
			Eventually(session.Out).Should(Say("cf auth USERNAME PASSWORD\n\n"))

			Eventually(session.Out).Should(Say("WARNING:"))
			Eventually(session.Out).Should(Say("Providing your password as a command line option is highly discouraged"))
			Eventually(session.Out).Should(Say("Your password may be visible to others and may be recorded in your shell history\n\n"))

			Eventually(session.Out).Should(Say("EXAMPLES:"))
			Eventually(session.Out).Should(Say("cf auth name@example\\.com \"my password\" \\(use quotes for passwords with a space\\)"))
			Eventually(session.Out).Should(Say("cf auth name@example\\.com \\\"\\\\\"password\\\\\"\\\" \\(escape quotes if used in password\\)\n\n"))

			Eventually(session.Out).Should(Say("SEE ALSO:"))
			Eventually(session.Out).Should(Say("api, login, target"))

			Eventually(session).Should(Exit(0))
		})
	})

	Context("when no arguments are provided", func() {
		It("errors-out with the help information", func() {
			session := helpers.CF("auth")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `USERNAME` and `PASSWORD` were not provided\n\n"))
			Eventually(session.Out).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	Context("when only a username is provided", func() {
		It("errors-out with a password required error and the help information", func() {
			session := helpers.CF("auth", "some-user")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `PASSWORD` was not provided\n\n"))
			Eventually(session.Out).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	Context("when too many arguments are provided", func() {
		It("displays an 'unknown flag' error message", func() {
			session := helpers.CF("auth", "some-username", "some-password", "-a", "api.bosh-lite.com")

			Eventually(session.Err).Should(Say("Incorrect Usage: unknown flag `a'"))
			Eventually(session.Out).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the API endpoint is not set", func() {
		BeforeEach(func() {
			helpers.UnsetAPI()
		})

		It("displays an error message", func() {
			session := helpers.CF("auth", "some-username", "some-password")

			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))

			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the user provides an invalid username/password combo", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
		})

		It("clears the cached tokens and target info, then displays an error message", func() {
			session := helpers.CF("auth", "some-username", "some-password")

			Eventually(session.Out).Should(Say("API endpoint: %s", helpers.GetAPI()))
			Eventually(session.Out).Should(Say("Authenticating\\.\\.\\."))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Credentials were rejected, please try again\\."))
			Eventually(session).Should(Exit(1))

			// Verify that the user is not logged-in
			targetSession1 := helpers.CF("target")
			Eventually(targetSession1.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
			Eventually(targetSession1.Out).Should(Say("FAILED"))
			Eventually(targetSession1).Should(Exit(1))

			// Verify that neither org nor space is targeted
			helpers.LoginCF()
			targetSession2 := helpers.CF("target")
			Eventually(targetSession2.Out).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
			Eventually(targetSession2).Should(Exit(0))
		})
	})

	Context("when the username and password are valid", func() {
		It("authenticates the user", func() {
			username, password := helpers.GetCredentials()
			session := helpers.CF("auth", username, password)

			Eventually(session.Out).Should(Say("API endpoint: %s", helpers.GetAPI()))
			Eventually(session.Out).Should(Say("Authenticating\\.\\.\\."))
			Eventually(session.Out).Should(Say("OK"))
			Eventually(session.Out).Should(Say("Use 'cf target' to view or set your target org and space"))

			Eventually(session).Should(Exit(0))
		})
	})
})
