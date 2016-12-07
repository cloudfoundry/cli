package integration

import (
	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-user command", func() {
	BeforeEach(func() {
		Skip("skipping until #126256695 is done")
		setAPI()
	})

	Context("when the environment is not setup correctly", func() {

		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				unsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := CF("create-user", PrefixedRandomName("integration-user"), PrefixedRandomName("password"))
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				logoutCF()
			})

			It("fails with not logged in message", func() {
				session := CF("create-user", PrefixedRandomName("integration-user"), PrefixedRandomName("password"))
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		Context("when logged-in user is not admin", func() {
			BeforeEach(func() {
				loginCF()
				noobUser := PrefixedRandomName("integration-user")
				noobPassword := PrefixedRandomName("password")
				session := CF("create-user", noobUser, noobPassword)
				Eventually(session).Should(Exit(0))
				session = CF("auth", noobUser, noobPassword)
				Eventually(session).Should(Exit(0))
			})

			It("fails with insufficient scope error", func() {
				session := CF("create-user", "user", "password")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Error Type: insufficient_scope"))
				Expect(session.Err).To(Say("Description: Insufficient scope for this resource"))
			})
		})
	})
})
