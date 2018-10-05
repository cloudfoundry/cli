package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("disable-feature-flags command", func() {
	BeforeEach(func() {
		helpers.LoginCF()
	})

	AfterEach(func() {
		session := helpers.CF("enable-feature-flag", "private_domain_creation")
		Eventually(session).Should(Exit(0))
	})

	It("disables a feature flag", func() {
		session := helpers.CF("disable-feature-flag", "private_domain_creation")
		Eventually(session).Should(Say("Setting status of private_domain_creation as"))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("feature-flag", "private_domain_creation")
		Eventually(session).Should(Say("private_domain_creation\\s+disabled"))
		Eventually(session).Should(Exit(0))
	})
})
