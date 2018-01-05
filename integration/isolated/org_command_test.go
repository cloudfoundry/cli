package isolated

import (
	"sort"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("org command", func() {
	var (
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("org", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("org - Show org info"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf org ORG [--guid]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--guid\\s+Retrieve and display the given org's guid.  All other output for the org is suppressed."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org-users, orgs"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "org", "org-name")
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		Context("when the org does not exist", func() {
			It("displays org not found and exits 1", func() {
				session := helpers.CF("org", orgName)
				userName, _ := helpers.GetCredentials()
				Eventually(session.Out).Should(Say("Getting info for org %s as %s\\.\\.\\.", orgName, userName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Organization '%s' not found.", orgName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the org exists", func() {
			BeforeEach(func() {
				setupCF(orgName, spaceName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			Context("when the --guid flag is used", func() {
				It("displays the org guid", func() {
					session := helpers.CF("org", "--guid", orgName)
					Eventually(session.Out).Should(Say("[\\da-f]{8}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{12}"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when no flags are used", func() {
				var (
					domainName              string
					quotaName               string
					spaceName2              string
					isolationSegmentsSorted []string
				)

				BeforeEach(func() {
					domainName = helpers.DomainName("")
					domain := helpers.NewDomain(orgName, domainName)
					domain.Create()

					quotaName = helpers.QuotaName()
					session := helpers.CF("create-quota", quotaName)
					Eventually(session).Should(Exit(0))
					session = helpers.CF("set-quota", orgName, quotaName)
					Eventually(session).Should(Exit(0))

					spaceName2 = helpers.NewSpaceName()
					helpers.CreateSpace(spaceName2)

					isolationSegmentName1 := helpers.NewIsolationSegmentName()
					Eventually(helpers.CF("create-isolation-segment", isolationSegmentName1)).Should(Exit(0))
					Eventually(helpers.CF("enable-org-isolation", orgName, isolationSegmentName1)).Should(Exit(0))

					isolationSegmentName2 := helpers.NewIsolationSegmentName()
					Eventually(helpers.CF("create-isolation-segment", isolationSegmentName2)).Should(Exit(0))
					Eventually(helpers.CF("enable-org-isolation", orgName, isolationSegmentName2)).Should(Exit(0))

					isolationSegmentsSorted = []string{isolationSegmentName1, isolationSegmentName2}
					sort.Strings(isolationSegmentsSorted)

					Eventually(helpers.CF("set-org-default-isolation-segment", orgName, isolationSegmentsSorted[0])).Should(Exit(0))
				})

				It("displays a table with org domains, quotas, spaces, space quotas and isolation segments, and exits 0", func() {
					session := helpers.CF("org", orgName)
					userName, _ := helpers.GetCredentials()
					Eventually(session.Out).Should(Say("Getting info for org %s as %s\\.\\.\\.", orgName, userName))

					Eventually(session.Out).Should(Say("name:\\s+%s", orgName))

					domainsSorted := []string{defaultSharedDomain(), domainName}
					sort.Strings(domainsSorted)
					Eventually(session.Out).Should(Say("domains:.+%s,.+%s", domainsSorted[0], domainsSorted[1]))

					Eventually(session.Out).Should(Say("quota:\\s+%s", quotaName))

					spacesSorted := []string{spaceName, spaceName2}
					sort.Strings(spacesSorted)
					Eventually(session.Out).Should(Say("spaces:\\s+%s,.* %s", spacesSorted[0], spacesSorted[1]))

					Eventually(session.Out).Should(Say("isolation segments:\\s+.*%s \\(default\\),.* %s", isolationSegmentsSorted[0], isolationSegmentsSorted[1]))

					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
