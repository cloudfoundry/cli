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

var _ = Describe("map-route command", func() {
	Context("help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("map-route", "ROUTES", "Map a route to an app"))
		})

		It("displays command usage to output", func() {
			session := helpers.CF("map-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`map-route - Map a route to an app\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`Map an HTTP route:\n`))
			Eventually(session).Should(Say(`cf map-route APP_NAME DOMAIN \[--hostname HOSTNAME\] \[--path PATH\] \[--app-protocol PROTOCOL\]\n`))
			Eventually(session).Should(Say(`Map a TCP route:\n`))
			Eventually(session).Should(Say(`cf map-route APP_NAME DOMAIN \[--port PORT]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf map-route my-app example.com                                                # example.com`))
			Eventually(session).Should(Say(`cf map-route my-app example.com --hostname myhost                              # myhost.example.com`))
			Eventually(session).Should(Say(`cf map-route my-app example.com --hostname myhost --path foo                   # myhost.example.com/foo`))
			Eventually(session).Should(Say(`cf map-route my-app example.com --hostname myhost --app-protocol http2 # myhost.example.com`))
			Eventually(session).Should(Say(`cf map-route my-app example.com --port 5000                                    # example.com:5000`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--path\s+Path for the HTTP route`))
			Eventually(session).Should(Say(`--port\s+Port for the TCP route \(default: random port\)`))
			Eventually(session).Should(Say(`--app-protocol\s+\[Beta flag, subject to change\] Protocol for the route destination \(default: http1\). Only applied to HTTP routes`))

			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`create-route, routes, unmap-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "map-route", "app-name", "domain-name")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName    string
			spaceName  string
			domainName string
			hostName   string
			path       string
			userName   string
			appName    string
		)

		BeforeEach(func() {
			appName = helpers.NewAppName()
			hostName = helpers.NewHostName()
			path = helpers.NewPath()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()

			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, "push", appName)
				Eventually(session).Should(Exit(0))
			})
		})

		When("the http route already exists", func() {
			var (
				route helpers.Route
			)

			BeforeEach(func() {
				domainName = helpers.DefaultSharedDomain()
				route = helpers.NewRoute(spaceName, domainName, hostName, path)
				route.V7Create()
			})

			AfterEach(func() {
				route.Delete()
			})
			When("route is already mapped to app", func() {
				BeforeEach(func() {
					session := helpers.CF("map-route", appName, domainName, "--hostname", route.Host, "--path", route.Path)
					Eventually(session).Should(Exit(0))
				})
				It("exits 0 with helpful message saying that the route is already mapped to the app", func() {
					session := helpers.CF("map-route", appName, domainName, "--hostname", route.Host, "--path", route.Path)

					Eventually(session).Should(Say(`Mapping route %s.%s%s to app %s in org %s / space %s as %s\.\.\.`, hostName, domainName, path, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`App '%s' is already mapped to route '%s.%s%s'\. Nothing has been updated\.`, appName, hostName, domainName, path))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))

				})
			})
			When("route is not yet mapped to the app", func() {
				It("maps the route to an app", func() {
					session := helpers.CF("map-route", appName, domainName, "--hostname", route.Host, "--path", route.Path)

					Eventually(session).Should(Say(`Mapping route %s.%s%s to app %s in org %s / space %s as %s\.\.\.`, hostName, domainName, path, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})

				When("destination protocol is provided", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionHTTP2RoutingV3)
					})

					It("maps the route to an app", func() {
						session := helpers.CF("map-route", appName, domainName, "--hostname", route.Host, "--app-protocol", "http2")

						Eventually(session).Should(Say(`Mapping route %s.%s to app %s with protocol http2 in org %s / space %s as %s\.\.\.`, hostName, domainName, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		When("the tcp route already exists", func() {
			var (
				domain      helpers.Domain
				routerGroup helpers.RouterGroup
				route       helpers.Route
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName()
				domain = helpers.NewDomain("", domainName)
				routerGroup = helpers.NewRouterGroup(
					helpers.NewRouterGroupName(),
					"1024-2048",
				)

				routerGroup.Create()
				domain.CreateWithRouterGroup(routerGroup.Name)
				route = helpers.NewTCPRoute(spaceName, domainName, 1082)
			})

			AfterEach(func() {
				domain.DeleteShared()
				routerGroup.Delete()
			})

			When("route is already mapped to app", func() {
				BeforeEach(func() {
					session := helpers.CF("map-route", appName, domainName, "--port", fmt.Sprintf("%d", route.Port))
					Eventually(session).Should(Exit(0))
				})
				It("exits 0 with helpful message saying that the route is already mapped to the app", func() {
					session := helpers.CF("map-route", appName, domainName, "--port", fmt.Sprintf("%d", route.Port))

					Eventually(session).Should(Say(`Mapping route %s:%d to app %s in org %s / space %s as %s\.\.\.`, domainName, route.Port, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`App '%s' is already mapped to route '%s:%d'\. Nothing has been updated\.`, appName, domainName, route.Port))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("route is not yet mapped to the app", func() {
				It("maps the route to an app", func() {
					session := helpers.CF("map-route", appName, domainName, "--port", fmt.Sprintf("%d", route.Port))

					Eventually(session).Should(Say(`Mapping route %s:%d to app %s in org %s / space %s as %s\.\.\.`, domainName, route.Port, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("port is not specified", func() {
				It("creates a new route with a random port and maps it to the app", func() {
					session := helpers.CF("map-route", appName, domainName)

					Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Say(`Mapping route %s:[0-9]+ to app %s in org %s / space %s as %s\.\.\.`, domainName, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the route does *not* exist", func() {
			When("it is an HTTP domain", func() {
				BeforeEach(func() {
					domainName = helpers.DefaultSharedDomain()
				})

				It("creates the route and maps it to an app", func() {
					session := helpers.CF("map-route", appName, domainName, "--hostname", hostName, "--path", path)
					Eventually(session).Should(Say(`Creating route %s.%s%s for org %s / space %s as %s\.\.\.`, hostName, domainName, path, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Say(`Mapping route %s.%s%s to app %s in org %s / space %s as %s\.\.\.`, hostName, domainName, path, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})

				When("destination protocol is provided", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionHTTP2RoutingV3)
					})

					It("maps the route to an app", func() {
						session := helpers.CF("map-route", appName, domainName, "--hostname", hostName, "--app-protocol", "http2")
						Eventually(session).Should(Say(`Creating route %s.%s for org %s / space %s as %s\.\.\.`, hostName, domainName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Say(`Mapping route %s.%s to app %s with protocol http2 in org %s / space %s as %s\.\.\.`, hostName, domainName, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("it is an TCP domain", func() {
				var (
					domain      helpers.Domain
					routerGroup helpers.RouterGroup
				)

				BeforeEach(func() {
					domainName = helpers.NewDomainName()
					domain = helpers.NewDomain("", domainName)
					routerGroup = helpers.NewRouterGroup(
						helpers.NewRouterGroupName(),
						"1024-2048",
					)

					routerGroup.Create()
					domain.CreateWithRouterGroup(routerGroup.Name)
				})

				AfterEach(func() {
					domain.DeleteShared()
					routerGroup.Delete()
				})

				When("a port is provided", func() {
					It("creates the route and maps it to an app", func() {
						session := helpers.CF("map-route", appName, domainName, "--port", "1052")
						Eventually(session).Should(Say(`Creating route %s:%s for org %s / space %s as %s\.\.\.`, domainName, "1052", orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Say(`Mapping route %s:%s to app %s in org %s / space %s as %s\.\.\.`, domainName, "1052", appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("a port is not provided", func() {
					It("creates the route and maps it to an app", func() {
						session := helpers.CF("map-route", appName, domainName)
						Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Say(`Mapping route %s:[0-9]+ to app %s in org %s / space %s as %s\.\.\.`, domainName, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
