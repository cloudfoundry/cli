package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("routes command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("routes", "ROUTES", "List all routes in the current space or the current organization"))
		})

		It("displays the help information", func() {
			session := helpers.CF("routes", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`routes - List all routes in the current space or the current organization\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf routes \[--org-level\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--org-level\s+List all the routes for all spaces of current organization`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`check-route, create-route, domains, map-route, unmap-route`))

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
			userName  string
			appName1  string
			appName2  string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("routes exist", func() {
			var (
				domainName string
				domain     helpers.Domain

				otherSpaceName string
			)

			BeforeEach(func() {
				otherSpaceName = helpers.GenerateHigherName(helpers.NewSpaceName, spaceName)

				domainName = helpers.NewDomainName()

				domain = helpers.NewDomain(orgName, domainName)
				appName1 = helpers.NewAppName()
				Eventually(helpers.CF("create-app", appName1)).Should(Exit(0))
				domain.CreatePrivate()
				Eventually(helpers.CF("map-route", appName1, domainName, "--hostname", "route1")).Should(Exit(0))
				Eventually(helpers.CF("map-route", appName1, domainName, "--hostname", "route1a")).Should(Exit(0))
				Eventually(helpers.CF("map-route", appName1, domainName, "--hostname", "route1b")).Should(Exit(0))
				Eventually(helpers.CF("set-label", "route", "route1b."+domainName, "env=prod")).Should(Exit(0))

				helpers.SetupCF(orgName, otherSpaceName)
				appName2 = helpers.NewAppName()
				Eventually(helpers.CF("create-app", appName2)).Should(Exit(0))
				Eventually(helpers.CF("map-route", appName2, domainName, "--hostname", "route2", "--path", "dodo")).Should(Exit(0))

				helpers.SetupCF(orgName, spaceName)
			})

			AfterEach(func() {
				domain.Delete()
				helpers.QuickDeleteSpace(otherSpaceName)
			})

			It("lists all the routes", func() {
				session := helpers.CF("routes")
				Eventually(session).Should(Exit(0))

				Expect(session).To(Say(`Getting routes for org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Expect(session).To(Say(`space\s+host\s+domain\s+path\s+apps`))
				Expect(session).To(Say(`%s\s+route1\s+%s\s+%s`, spaceName, domainName, appName1))
				Expect(session).To(Say(`%s\s+route1a\s+%s\s+%s`, spaceName, domainName, appName1))
				Expect(session).To(Say(`%s\s+route1b\s+%s\s+%s`, spaceName, domainName, appName1))
				Expect(session).ToNot(Say(`%s\s+route2\s+%s\s+%s`, spaceName, domainName, appName2))
			})

			It("lists all the routes by label", func() {
				session := helpers.CF("routes", "--labels", "env in (prod)")
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say(`Getting routes for org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Expect(session).To(Say(`space\s+host\s+domain\s+path\s+apps`))
				Expect(session).ToNot(Say(`%s\s+route1\s+%s\s+%s`, spaceName, domainName, appName1))
				Expect(session).ToNot(Say(`%s\s+route1a\s+%s\s+%s`, spaceName, domainName, appName1))
				Expect(session).To(Say(`%s\s+route1b\s+%s\s+%s`, spaceName, domainName, appName1))
				Expect(session).ToNot(Say(`%s\s+route2\s+%s\s+%s`, spaceName, domainName, appName2))
			})

			When("fetching routes by org", func() {
				It("lists all the routes in the org", func() {
					session := helpers.CF("routes", "--org-level")
					Eventually(session).Should(Say(`Getting routes for org %s as %s\.\.\.`, orgName, userName))
					Eventually(session).Should(Say(`space\s+host\s+domain\s+path\s+apps`))

					Eventually(session).Should(Say(`%s\s+route1\s+%s\s+%s`, spaceName, domainName, appName1))
					Eventually(session).Should(Say(`%s\s+route2\s+%s\s+\/dodo\s+%s`, otherSpaceName, domainName, appName2))
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
