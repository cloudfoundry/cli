package isolated

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update destination command", func() {
	Context("Heltp", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")

			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("update-destination", "ROUTES", "Updates the destination protocol for a route"))
		})

		It("displays the help information", func() {
			session := helpers.CF("update-destination", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`update-destination - Updates the destination protocol for a route`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`Edit an existing HTTP route`))
			Eventually(session).Should(Say(`cf update-destination APP_NAME DOMAIN \[--hostname HOSTNAME\] \[--app-protocol PROTOCOL\] \[--path PATH\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf update-destination my-app example.com --hostname myhost --app-protocol http2                   # myhost.example.com`))
			Eventually(session).Should(Say(`cf update destination my-app example.com --hostname myhost --path foo --app-protocol http2        # myhost.example.com/foo`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--app-protocol\s+New Protocol for the route destination \(http1 or http2\). Only applied to HTTP routes`))
			Eventually(session).Should(Say(`--path\s+Path for the HTTP route`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`create-route, map-route, routes, unmap-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "update-destination", "app-name", "some-domain")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			userName  string
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionHTTP2RoutingV3)
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
					appName  string
					hostname string
				)

				When("it's an HTTP/1 route", func() {
					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						hostname = "key-lime-pie"
						appName = "killer"
						domain.CreatePrivate()
						Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
						Eventually(helpers.CF("create-route", domain.Name, "--hostname", hostname)).Should(Exit(0))
						Eventually(helpers.CF("map-route", appName, domain.Name, "--hostname", hostname)).Should(Exit(0))
					})

					AfterEach(func() {
						domain.Delete()
					})

					It("updates the destination protocol to http2 ", func() {
						session := helpers.CF("update-destination", appName, domainName, "--hostname", hostname, "--app-protocol", "http2")
						Eventually(session).Should(Say(`Updating destination protocol from %s to %s for route %s.%s in org %s / space %s as %s...`, "http1", "http2", hostname, domainName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("it's an HTTP/2 route", func() {
					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						hostname = "key-lime-pie"
						appName = "killer2"
						domain.CreatePrivate()
						Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
						Eventually(helpers.CF("create-route", domain.Name, "--hostname", hostname)).Should(Exit(0))
						Eventually(helpers.CF("map-route", appName, domain.Name, "--hostname", hostname, "--app-protocol", "http2")).Should(Exit(0))
					})

					AfterEach(func() {
						domain.Delete()
					})

					It("updates the destination protocol to http1 ", func() {
						session := helpers.CF("update-destination", appName, domainName, "--hostname", hostname, "--app-protocol", "http1")
						Eventually(session).Should(Say(`Updating destination protocol from %s to %s for route %s.%s in org %s / space %s as %s...`, "http2", "http1", hostname, domainName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})

					It("does nothing when the app-protocol is the same", func() {
						session := helpers.CF("update-destination", appName, domainName, "--hostname", hostname, "--app-protocol", "http2")
						Eventually(session).Should(Say(`App '%s' is already using '%s'\. Nothing has been updated`, appName, "http2"))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})

					It("sets the destination protocol to 'http1' when the app-protocol is not provided", func() {
						session := helpers.CF("update-destination", appName, domainName, "--hostname", hostname)
						Eventually(session).Should(Say(`Updating destination protocol from %s to %s for route %s.%s in org %s / space %s as %s...`, "http2", "http1", hostname, domainName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
