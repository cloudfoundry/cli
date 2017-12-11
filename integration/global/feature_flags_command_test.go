package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("feature-flags command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("feature-flags", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("\\s+feature-flags - Retrieve list of feature flags with status"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("\\s+cf feature-flags"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("\\s+disable-feature-flag, enable-feature-flag"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the API endpoint is not set", func() {
		BeforeEach(func() {
			helpers.UnsetAPI()
		})

		It("displays an error message and exits 1", func() {
			session := helpers.CF("feature-flags")
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays an error and exits 1", func() {
			session := helpers.CF("feature-flags")
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the API endpoint is set and the user is logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("displays a list of feature flags with current state and exits 0", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("feature-flags")
			Eventually(session.Out).Should(Say("Retrieving status of all flagged features as %s\\.\\.\\.", username))
			Eventually(session.Out).Should(Say(""))
			Eventually(session.Out).Should(Say("features\\s+state"))
			Eventually(session.Out).Should(Say("[a-z_]+\\s+(enabled|disabled)"))
			Eventually(session).Should(Exit(0))
		})
	})
})
