package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("feature-flags command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("feature-flags", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+feature-flags - Retrieve list of feature flags with status`))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf feature-flags`))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+disable-feature-flag, enable-feature-flag`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the API endpoint is not set", func() {
		BeforeEach(func() {
			helpers.UnsetAPI()
		})

		It("displays an error message and exits 1", func() {
			session := helpers.CF("feature-flags")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`No API endpoint set\. Use 'cf login' or 'cf api' to target an endpoint\.`))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays an error and exits 1", func() {
			session := helpers.CF("feature-flags")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`Not logged in\. Use 'cf login' or 'cf login --sso' to log in\.`))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the API endpoint is set and the user is logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("displays a list of feature flags with current state and exits 0", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("feature-flags")
			Eventually(session).Should(Say(`Retrieving status of all flagged features as %s\.\.\.`, username))
			Eventually(session).Should(Say(""))
			Eventually(session).Should(Say(`features\s+state`))
			Eventually(session).Should(Say(`[a-z_]+\s+(enabled|disabled)`))
			Eventually(session).Should(Exit(0))
		})
	})
})
