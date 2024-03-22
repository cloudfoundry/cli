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

var _ = Describe("route command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("route", "ROUTES", "Display route details and mapped destinations"))
		})

		It("displays the help information", func() {
			session := helpers.CF("route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`route - Display route details and mapped destinations`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`Display an HTTP route:`))
			Eventually(session).Should(Say(`cf route DOMAIN \[--hostname HOSTNAME\] \[--path PATH\]\n`))
			Eventually(session).Should(Say(`Display a TCP route:`))
			Eventually(session).Should(Say(`cf route DOMAIN --port PORT\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf route example.com                      # example.com`))
			Eventually(session).Should(Say(`cf route example.com -n myhost --path foo # myhost.example.com/foo`))
			Eventually(session).Should(Say(`cf route example.com --path foo           # example.com/foo`))
			Eventually(session).Should(Say(`cf route example.com --port 5000          # example.com:5000`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname used to identify the HTTP route`))
			Eventually(session).Should(Say(`--path\s+Path used to identify the HTTP route`))
			Eventually(session).Should(Say(`--port\s+Port used to identify the TCP route`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`create-route, delete-route, routes`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "route", "some-domain")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			userName  string
			orgName   string
			spaceName string
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

		When("the domain exists", func() {
			var (
				domainName string
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName()
			})

			When("the route exists", func() {
				var (
					domain   helpers.Domain
					hostname string
					path     string
				)

				When("it's an HTTP route", func() {
					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						hostname = "key-lime-pie"
						path = "/some-path"
						domain.CreatePrivate()
						Eventually(helpers.CF("create-app", "killer")).Should(Exit(0))
						Eventually(helpers.CF("create-route", domain.Name, "--hostname", hostname, "--path", path)).Should(Exit(0))
						Eventually(helpers.CF("map-route", "killer", domain.Name, "--hostname", hostname, "--path", path)).Should(Exit(0))
					})

					AfterEach(func() {
						domain.Delete()
					})

					It("displays the route summary and exits without failing", func() {
						session := helpers.CF("route", domainName, "--hostname", hostname, "--path", path)
						Eventually(session).Should(Say(`Showing route %s\.%s%s in org %s / space %s as %s\.\.\.`, hostname, domainName, path, orgName, spaceName, userName))
						Eventually(session).Should(Say(`domain:\s+%s`, domainName))
						Eventually(session).Should(Say(`host:\s+%s`, hostname))
						Eventually(session).Should(Say(`port:\s+\n`))
						Eventually(session).Should(Say(`path:\s+%s`, path))
						Eventually(session).Should(Say(`protocol:\s+http`))
						Eventually(session).Should(Say(`\n`))
						Eventually(session).Should(Say(`Destinations:`))
						Eventually(session).Should(Say(`\s+app\s+process\s+port\s+app-protocol`))
						Eventually(session).Should(Exit(0))
					})

					It("displays the protocol in the route summary and exits without failing", func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionHTTP2RoutingV3)
						session := helpers.CF("route", domainName, "--hostname", hostname, "--path", path)
						Eventually(session).Should(Say(`\s+killer\s+web\s+8080\s+http1`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("it's a TCP route", func() {
					var (
						routerGroup helpers.RouterGroup
						port        int
						tcpDomain   helpers.Domain
					)

					BeforeEach(func() {
						routerGroup = helpers.NewRouterGroup(helpers.NewRouterGroupName(), "1024-2048")
						routerGroup.Create()

						tcpDomain = helpers.NewDomain(orgName, helpers.NewDomainName("TCP-DOMAIN"))
						tcpDomain.CreateWithRouterGroup(routerGroup.Name)

						port = 1024

						Eventually(helpers.CF("create-app", "killer")).Should(Exit(0))
						Eventually(helpers.CF("create-route", tcpDomain.Name, "--port", fmt.Sprintf("%d", port))).Should(Exit(0))
						Eventually(helpers.CF("map-route", "killer", tcpDomain.Name, "--port", "1024")).Should(Exit(0))
					})

					AfterEach(func() {
						tcpDomain.DeleteShared()
						routerGroup.Delete()
					})

					It("displays the route summary and exits without failing", func() {
						session := helpers.CF("route", tcpDomain.Name, "--port", fmt.Sprintf("%d", port))
						Eventually(session).Should(Say(`Showing route %s:%d in org %s / space %s as %s\.\.\.`, tcpDomain.Name, port, orgName, spaceName, userName))
						Eventually(session).Should(Say(`domain:\s+%s`, tcpDomain.Name))
						Eventually(session).Should(Say(`host:\s+\n`))
						Eventually(session).Should(Say(`port:\s+%d`, port))
						Eventually(session).Should(Say(`path:\s+\n`))
						Eventually(session).Should(Say(`protocol:\s+tcp`))
						Eventually(session).Should(Say(`\n`))
						Eventually(session).Should(Say(`Destinations:`))
						Eventually(session).Should(Say(`\s+app\s+process\s+port\s+app-protocol`))
						Eventually(session).Should(Exit(0))
					})

					It("displays the protocol in the route summary and exits without failing", func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionHTTP2RoutingV3)
						session := helpers.CF("route", tcpDomain.Name, "--port", fmt.Sprintf("%d", port))
						Eventually(session).Should(Say(`\s+killer\s+web\s+8080\s+tcp`))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the route does not exist", func() {
				var domain helpers.Domain

				BeforeEach(func() {
					domain = helpers.NewDomain(orgName, domainName)
					domain.Create()
				})

				AfterEach(func() {
					domain.Delete()
				})

				When("no flags are used", func() {
					It("checks the route", func() {
						session := helpers.CF("route", domainName)
						Eventually(session).Should(Say(`Showing route %s in org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
						Eventually(session.Err).Should(Say(`Route with host '', domain '%s', and path '/' not found\.`, domainName))
						Eventually(session).Should(Exit(1))
					})
				})

				When("passing in a hostname", func() {
					It("checks the route with the hostname", func() {
						hostname := "tiramisu"
						session := helpers.CF("route", domainName, "-n", hostname)
						Eventually(session).Should(Say(`Showing route %s.%s in org %s / space %s as %s\.\.\.`, hostname, domainName, orgName, spaceName, userName))
						Eventually(session.Err).Should(Say(`Route with host '%s', domain '%s', and path '/' not found\.`, hostname, domainName))
						Eventually(session).Should(Exit(1))
					})
				})

				When("passing in hostname and path with a leading '/'", func() {
					It("checks the route with hostname and path", func() {
						hostname := "tiramisu"
						pathString := "/recipes"
						session := helpers.CF("route", domainName, "-n", hostname, "--path", pathString)
						Eventually(session).Should(Say(`Showing route %s.%s%s in org %s / space %s as %s\.\.\.`, hostname, domainName, pathString, orgName, spaceName, userName))
						Eventually(session.Err).Should(Say(`Route with host '%s', domain '%s', and path '%s' not found`, hostname, domainName, pathString))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		When("the domain does not exist", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("route", "some-domain")
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Domain 'some-domain' not found.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the domain is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("route")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `DOMAIN` was not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
