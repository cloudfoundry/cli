package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("routes command", func() {

	appProtocolValue := "http1"
	const tableHeaders = `space\s+host\s+domain\s+port\s+path\s+protocol\s+app-protocol\s+apps\s+service instance`
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
			if !helpers.IsVersionMet(ccversion.MinVersionHTTP2RoutingV3) {
				appProtocolValue = ""
			}

		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("routes exist", func() {
			var (
				domainName string
				domain     helpers.Domain

				otherSpaceName      string
				serviceInstanceName string
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

				serviceInstanceName = helpers.NewServiceInstanceName()
				routeServiceURL := helpers.RandomURL()
				Eventually(helpers.CF("cups", serviceInstanceName, "-r", routeServiceURL)).Should(Exit(0))
				Eventually(helpers.CF("bind-route-service", domainName, "--hostname", "route1", serviceInstanceName, "--wait")).Should(Exit(0))

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
				Expect(session).To(Say(tableHeaders))
				Expect(session).To(Say(`%s\s+route1\s+%s\s+http\s+%s\s+%s\s+%s\n`, spaceName, domainName, appProtocolValue, appName1, serviceInstanceName))
				Expect(session).To(Say(`%s\s+route1a\s+%s\s+http\s+%s\s+%s\s+\n`, spaceName, domainName, appProtocolValue, appName1))
				Expect(session).To(Say(`%s\s+route1b\s+%s\s+http\s+%s\s+%s\s+\n`, spaceName, domainName, appProtocolValue, appName1))
				Expect(session).ToNot(Say(`%s\s+route2\s+%s\s+http\s+%s\s+%s\s+\n`, spaceName, domainName, appProtocolValue, appName2))
			})

			It("lists all the routes by label", func() {
				session := helpers.CF("routes", "--labels", "env in (prod)")
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say(`Getting routes for org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Expect(session).To(Say(tableHeaders))
				Expect(session).ToNot(Say(`%s\s+route1\s+%s\s+http\s+%s\s+%s\s+%s\n`, spaceName, domainName, appProtocolValue, appName1, serviceInstanceName))
				Expect(session).ToNot(Say(`%s\s+route1a\s+%s\s+http\s+%s\s+%s\s+\n`, spaceName, domainName, appProtocolValue, appName1))
				Expect(session).To(Say(`%s\s+route1b\s+%s\s+http\s+%s\s+%s\s+\n`, spaceName, domainName, appProtocolValue, appName1))
				Expect(session).ToNot(Say(`%s\s+route2\s+%s\s+http\s+%s\s+%s\s+\n`, spaceName, domainName, appProtocolValue, appName2))
			})

			When("fetching routes by org", func() {
				It("lists all the routes in the org", func() {
					session := helpers.CF("routes", "--org-level")
					Eventually(session).Should(Exit(0))
					Expect(session).To(Say(`Getting routes for org %s as %s\.\.\.`, orgName, userName))
					Expect(session).To(Say(tableHeaders))
					Expect(session).To(Say(`%s\s+route1\s+%s\s+http\s+%s\s+%s\s+%s\n`, spaceName, domainName, appProtocolValue, appName1, serviceInstanceName))
					Expect(session).To(Say(`%s\s+route2\s+%s\s+\/dodo\s+http\s+%s\s+%s\s+\n`, otherSpaceName, domainName, appProtocolValue, appName2))
				})
			})
		})

		When("http1 and http2 routes exist", func() {
			var (
				domainName string
				domain     helpers.Domain
			)

			BeforeEach(func() {
				helpers.SkipIfVersionLessThan(ccversion.MinVersionHTTP2RoutingV3)
				domainName = helpers.NewDomainName()

				domain = helpers.NewDomain(orgName, domainName)

				appName1 = helpers.NewAppName()
				Eventually(helpers.CF("create-app", appName1)).Should(Exit(0))
				appName2 = helpers.NewAppName()
				Eventually(helpers.CF("create-app", appName2)).Should(Exit(0))

				domain.CreatePrivate()

				Eventually(helpers.CF("map-route", appName1, domainName, "--hostname", "route1")).Should(Exit(0))
				Eventually(helpers.CF("map-route", appName2, domainName, "--hostname", "route2")).Should(Exit(0))
				Eventually(helpers.CF("map-route", appName2, domainName, "--hostname", "route1", "--app-protocol", "http2")).Should(Exit(0))

				helpers.SetupCF(orgName, spaceName)
			})

			AfterEach(func() {
				domain.Delete()
			})

			It("lists all the routes", func() {
				session := helpers.CF("routes")
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say(`Getting routes for org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Expect(session).To(Say(tableHeaders))
				Eventually(session).Should(Or(Say(`%s\s+route1\s+%s\s+http\s+http1, http2\s+%s\s+\n`, spaceName, domainName, fmt.Sprintf("%s, %s", appName2, appName1)), Say(`%s\s+route1\s+%s\s+http\s+http1, http2\s+%s\s+\n`, spaceName, domainName, fmt.Sprintf("%s, %s", appName1, appName2))))
				Eventually(session).Should(Say(`%s\s+route2\s+%s\s+http\s+http1\s+%s\s+\n`, spaceName, domainName, appName2))
			})
		})

		When("when shared tcp routes exist", func() {
			var (
				domainName  string
				domain      helpers.Domain
				routerGroup helpers.RouterGroup
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName()

				domain = helpers.NewDomain(orgName, domainName)

				routerGroup = helpers.NewRouterGroup(
					helpers.NewRouterGroupName(),
					"1024-2048",
				)
				routerGroup.Create()
				domain.CreateWithRouterGroup(routerGroup.Name)

				Eventually(helpers.CF("create-route", domainName, "--port", "1028")).Should(Exit(0))
			})

			AfterEach(func() {
				domain.DeleteShared()
				routerGroup.Delete()
			})

			It("lists all the routes", func() {
				session := helpers.CF("routes")
				Eventually(session).Should(Exit(0))

				Expect(session).To(Say(`Getting routes for org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Expect(session).To(Say(tableHeaders))
				Expect(session).To(Say(`%s\s+%s[^:]\s+%d\s+tcp`, spaceName, domainName, 1028))
			})
		})

		When("no routes exist", func() {
			It("outputs a message about no routes existing", func() {
				session := helpers.CF("routes")
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say(`Getting routes for org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Expect(session).To(Say("No routes found"))
			})
		})
	})
})
