package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("events command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app1")
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("events", "APPS", "Show recent app events"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("events", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("events - Show recent app events"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf events APP_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("app, logs, map-route, unmap-route"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "events", appName)
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("with an existing app", func() {
			BeforeEach(func() {
				session := helpers.CF("create-app", appName)
				Eventually(session).Should(Exit(0))
				session = helpers.CF("rename", appName, "other-app-name")
				Eventually(session).Should(Exit(0))
				session = helpers.CF("rename", "other-app-name", appName)
				Eventually(session).Should(Exit(0))
			})

			It("displays events in the list", func() {
				session := helpers.CF("events", appName)

				Eventually(session).Should(Say(`Getting events for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say(`time\s+event\s+actor\s+description`))
				Eventually(session).Should(Say(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{2}-\d{4}\s+audit\.app\.update\s+%s`, userName))
				Eventually(session).Should(Say(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{2}-\d{4}\s+audit\.app\.update\s+%s`, userName))
				Eventually(session).Should(Say(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{2}-\d{4}\s+audit\.app\.create\s+%s`, userName))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})
