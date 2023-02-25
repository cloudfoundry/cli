package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("feature-flags command", func() {
	When("--help flag is set", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("feature-flags", "FEATURE FLAGS", "Retrieve list of feature flags with status"))
		})

		It("Displays command usage to output", func() {
			session := helpers.CF("feature-flags", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`feature-flags - Retrieve list of feature flags with status`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf feature-flags`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`disable-feature-flag, enable-feature-flag, feature-flag`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("environment is not set up correctly", func() {
		It("displays an error and exits 1", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "feature-flags")
		})
	})

	When("environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("displays feature flag settings", func() {
			session := helpers.CF("feature-flags")
			username, _ := helpers.GetCredentials()
			Eventually(session).Should(Say(`Getting feature flags as %s\.\.\.`, username))
			Eventually(session).Should(Say(`name\s+state`))
			Eventually(session).Should(Say(`app_scaling\s+(dis|en)abled`))
			Eventually(session).Should(Say(`service_instance_creation\s+(dis|en)abled`))
			Eventually(session).Should(Say(`user_org_creation\s+(dis|en)abled`))
			Eventually(session).Should(Exit(0))
		})
	})
})
