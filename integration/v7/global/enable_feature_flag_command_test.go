package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("enable-feature-flags command", func() {

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("enable-feature-flag", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("enable-feature-flag - Allow use of a feature"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf enable-feature-flag FEATURE_FLAG_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("disable-feature-flag, feature-flag, feature-flags"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	BeforeEach(func() {
		helpers.LoginCF()
	})

	AfterEach(func() {
		session := helpers.CF("disable-feature-flag", "user_org_creation")
		Eventually(session).Should(Exit(0))
	})

	It("enables a feature flag", func() {
		session := helpers.CF("enable-feature-flag", "user_org_creation")
		Eventually(session).Should(Say("Enabling feature flag user_org_creation as"))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("feature-flag", "user_org_creation")
		Eventually(session).Should(Say(`user_org_creation\s+enabled`))
		Eventually(session).Should(Exit(0))
	})
})
