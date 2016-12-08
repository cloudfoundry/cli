package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("feature-flag command", func() {
	BeforeEach(func() {
		helpers.LoginCF()
	})

	It("displays feature flag settings", func() {
		session := helpers.CF("feature-flag", "user_org_creation")
		Eventually(session).Should(Say("Retrieving status of user_org_creation as"))
		Eventually(session).Should(Say("user_org_creation\\s+(dis|en)abled"))
		Eventually(session).Should(Exit(0))
	})
})
