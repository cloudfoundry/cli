package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unmap-route command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unmap-route", "--help")
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`unmap-route - Remove a url route from an app\n`))
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
				Eventually(session).Should(Say(`delete-route, routes`))

				Eventually(session).Should(Exit(0))
			})
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
			appName    = helpers.NewAppName()
			route helpers.Route
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			hostName   = helpers.NewHostName()
			path       = helpers.NewPath()
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
			domainName = helpers.DefaultSharedDomain()
			route = helpers.NewRoute(spaceName, domainName, hostName, path)
			route.V7Create()

			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, "push", appName)
				Eventually(session).Should(Exit(0))
				session = helpers.CF("map-route", appName, domainName, "--hostname", route.Host, "--path", route.Path)
				Eventually(session).Should(Exit(0))
			})
		})

		AfterEach(func() {
			route.Delete()
		})

		When("the route exists and is mapped to the app", func() {
			It("unmaps the route from the app", func() {
				session := helpers.CF("unmap-route", appName, domainName, "--hostname", route.Host, "--path", route.Path)

				Eventually(session).Should(Say(`Removing route %s.%s%s from app %s in org %s / space %s as %s\.\.\.`, hostName, domainName, path, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say(`OK`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the route doesnt exist", func() {
			It("creates the route and maps it to an app", func() {
				session := helpers.CF("unmap-route", appName, domainName, "--hostname", "test", "--path", "route")

				Eventually(session).Should(Say(`FAILED`))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
