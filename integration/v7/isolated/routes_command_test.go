package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("routes command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("routes", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`routes - List all routes in the current space or the current organization\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf routes \[--orglevel\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--orglevel\s+List all the routes for all spaces of current organization`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`check-route, domains, map-route, unmap-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "routes")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		var (
			userName string
		)

		BeforeEach(func() {
			userName, _ = helpers.GetCredentials()
		})

		When("routes exist", func() {
			var (
				userName   string
				domainName string
				domain     helpers.Domain

				otherSpaceName string
			)

			BeforeEach(func() {
				otherSpaceName = helpers.NewSpaceName()

				domainName = helpers.NewDomainName()
				userName, _ = helpers.GetCredentials()

				domain = helpers.NewDomain(orgName, domainName)
				domain.CreatePrivate()
				Eventually(helpers.CF("create-route", domainName, "--hostname", "route1")).Should(Exit(0))
				helpers.SetupCF(orgName, otherSpaceName)
				Eventually(helpers.CF("create-route", domainName, "--hostname", "route2")).Should(Exit(0))
				helpers.SetupCF(orgName, spaceName)
			})

			AfterEach(func() {
				domain.Delete()
				helpers.QuickDeleteSpace(otherSpaceName)
			})

			It("lists all the routes", func() {
				session := helpers.CF("routes")
				Eventually(session).Should(Say(`Getting routes for org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Eventually(session).Should(Say(`%s\s+route1\s+%s`, spaceName, domainName))
				Eventually(session).Should(Exit(0))
			})

			When("fetching routes by org", func() {
				It("lists all the routes in the org", func() {
					session := helpers.CF("routes", "--orglevel")
					Eventually(session).Should(Say(`Getting routes for org %s as %s\.\.\.`, orgName, userName))
					Eventually(session).Should(Say(`%s\s+route1\s+%s`, spaceName, domainName))
					Eventually(session).Should(Say(`%s\s+route2\s+%s`, otherSpaceName, domainName))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("no routes exist", func() {
			It("outputs a message about no routes existing", func() {
				session := helpers.CF("routes")
				Eventually(session).Should(Say(`Getting routes for org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Eventually(session).Should(Say("No routes found"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
