package isolated

import (
	"fmt"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unmap-route command", func() {
	Context("help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("unmap-route", "ROUTES", "Remove a route from an app"))
		})

		It("Displays command usage to output", func() {
			session := helpers.CF("unmap-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`unmap-route - Remove a route from an app\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`Unmap an HTTP route:`))
			Eventually(session).Should(Say(`\s+cf unmap-route APP_NAME DOMAIN \[--hostname HOSTNAME\] \[--path PATH\]\n`))
			Eventually(session).Should(Say(`Unmap a TCP route:`))
			Eventually(session).Should(Say(`\s+cf unmap-route APP_NAME DOMAIN --port PORT\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf unmap-route my-app example.com                              # example.com`))
			Eventually(session).Should(Say(`cf unmap-route my-app example.com --hostname myhost            # myhost.example.com`))
			Eventually(session).Should(Say(`cf unmap-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo`))
			Eventually(session).Should(Say(`cf unmap-route my-app example.com --port 5000                  # example.com:5000`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname used to identify the HTTP route`))
			Eventually(session).Should(Say(`--path\s+Path used to identify the HTTP route`))
			Eventually(session).Should(Say(`--port\s+Port used to identify the TCP route`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`delete-route, map-route, routes`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "map-route", "app-name", "domain-name")
		})
	})

	When("The environment is set up correctly", func() {
		var (
			orgName    string
			spaceName  string
			hostName   string
			path       string
			domainName string
			userName   string
			appName    string
			route      helpers.Route
			tcpRoute   helpers.Route
			port       int
			tcpDomain  helpers.Domain
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.NewAppName()
			hostName = helpers.NewHostName()
			path = helpers.NewPath()
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
			domainName = helpers.DefaultSharedDomain()

			routerGroupName := helpers.FindOrCreateTCPRouterGroup(4)
			tcpDomain = helpers.NewDomain(orgName, helpers.NewDomainName("TCP-DOMAIN"))
			tcpDomain.CreateWithRouterGroup(routerGroupName)

			route = helpers.NewRoute(spaceName, domainName, hostName, path)
			route.V7Create()

			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, "push", appName)
				Eventually(session).Should(Exit(0))
			})
		})

		AfterEach(func() {
			route.Delete()
		})

		When("the route exists and is mapped to the app", func() {
			When("it's an http route", func() {
				BeforeEach(func() {
					session := helpers.CF("map-route", appName, domainName, "--hostname", route.Host, "--path", route.Path)
					Eventually(session).Should(Exit(0))
				})

				It("unmaps the route from the app", func() {
					session := helpers.CF("unmap-route", appName, domainName, "--hostname", route.Host, "--path", route.Path)

					Eventually(session).Should(Say(`Removing route %s.%s%s from app %s in org %s / space %s as %s\.\.\.`, hostName, domainName, path, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})
			When("it's a TCP route", func() {
				BeforeEach(func() {
					port = 1024
					tcpRoute = helpers.NewTCPRoute(spaceName, tcpDomain.Name, port)
					session := helpers.CF("map-route", appName, tcpDomain.Name, "--port", fmt.Sprintf("%d", tcpRoute.Port))
					Eventually(session).Should(Exit(0))
				})

				It("unmaps the route from the app", func() {
					session := helpers.CF("unmap-route", appName, tcpDomain.Name, "--port", fmt.Sprintf("%d", tcpRoute.Port))

					Eventually(session).Should(Say(`Removing route %s:%d from app %s in org %s / space %s as %s\.\.\.`, tcpDomain.Name, tcpRoute.Port, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the route exists but is not mapped to the app", func() {
			It("prints a message and exits with status 0", func() {
				session := helpers.CF("unmap-route", appName, domainName, "--hostname", route.Host, "--path", route.Path)

				Eventually(session).Should(Say("Route to be unmapped is not currently mapped to the application."))
				Eventually(session).Should(Say(`OK`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the route doesn't exist", func() {
			It("fails with a helpful message", func() {
				session := helpers.CF("unmap-route", appName, domainName, "--hostname", "test", "--path", "does-not-exist")

				Eventually(session.Err).Should(Say(`Route with host 'test', domain '%s', and path '/does-not-exist' not found.`, domainName))
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
