package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
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

	When("--help flag is set", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("feature-flag", "FEATURE FLAGS", "Retrieve an individual feature flag with status"))
		})

		It("Displays command usage to output", func() {
			session := helpers.CF("feature-flag", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("feature-flag - Retrieve an individual feature flag with status"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf feature-flag FEATURE_FLAG_NAME"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("disable-feature-flag, enable-feature-flag, feature-flags"))

			Eventually(session).Should(Exit(0))
		})
	})

	It("displays feature flag settings", func() {
		session := helpers.CF("feature-flag", "user_org_creation")
		Eventually(session).Should(Say("Getting info for feature flag user_org_creation as"))
		Eventually(session).Should(Say(`user_org_creation\s+(dis|en)abled`))
		Eventually(session).Should(Exit(0))
	})
})
