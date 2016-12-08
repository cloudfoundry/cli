package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("enable-feature-flags command", func() {
	BeforeEach(func() {
		helpers.LoginCF()
	})

	AfterEach(func() {
		session := helpers.CF("disable-feature-flag", "user_org_creation")
		Eventually(session).Should(Exit(0))
	})

	It("enables a feature flag", func() {
		session := helpers.CF("enable-feature-flag", "user_org_creation")
		Eventually(session).Should(Say("Setting status of user_org_creation as"))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("feature-flag", "user_org_creation")
		Eventually(session).Should(Say("user_org_creation\\s+enabled"))
		Eventually(session).Should(Exit(0))
	})
})
