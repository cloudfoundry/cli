package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("disable-feature-flags command", func() {

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("disable-feature-flag", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("disable-feature-flag - Prevent use of a feature"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf disable-feature-flag FEATURE_FLAG_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("enable-feature-flag, feature-flag, feature-flags"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	BeforeEach(func() {
		helpers.LoginCF()
	})

	AfterEach(func() {
		session := helpers.CF("enable-feature-flag", "private_domain_creation")
		Eventually(session).Should(Exit(0))
	})

	It("enables a feature flag", func() {
		session := helpers.CF("disable-feature-flag", "private_domain_creation")
		Eventually(session).Should(Say("Disabling feature flag private_domain_creation as"))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Say("Feature flag private_domain_creation disabled"))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("feature-flag", "private_domain_creation")
		Eventually(session).Should(Say(`private_domain_creation\s+disabled`))
		Eventually(session).Should(Exit(0))
	})
})
