package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-route command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("create-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`create-route - Create a route for later use\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf create-route DOMAIN \[--hostname HOSTNAME\] \[--path PATH\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf create-route example.com\s+# example.com`))
			Eventually(session).Should(Say(`cf create-route example.com --hostname myapp\s+# myapp.example.com`))
			Eventually(session).Should(Say(`cf create-route example.com --hostname myapp --path foo\s+# myapp.example.com/foo`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--path\s+Path for the HTTP route`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`check-route, domains, map-route, routes, unmap route`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "create-route", "some-domain")
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

		When("the space and domain exist", func() {
			var (
				userName   string
				domainName string
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName()
				userName, _ = helpers.GetCredentials()
			})

			When("the route already exists", func() {
				var (
					domain   helpers.Domain
					hostname string
				)

				BeforeEach(func() {
					domain = helpers.NewDomain(orgName, domainName)
					hostname = "key-lime-pie"
					domain.CreatePrivate()
					Eventually(helpers.CF("create-route", domainName, "--hostname", hostname)).Should(Exit(0))
				})

				AfterEach(func() {
					domain.Delete()
				})

				It("warns the user that it has already been created and runs to completion without failing", func() {
					session := helpers.CF("create-route", domainName, "--hostname", hostname)
					Eventually(session).Should(Say(`Creating route %s\.%s for org %s / space %s as %s\.\.\.`, hostname, domainName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Route already exists with host '%s' for domain '%s'\.`, hostname, domainName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the route does not already exist", func() {
				When("the domain is private", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						domain.Create()
					})

					AfterEach(func() {
						domain.Delete()
					})

					When("no flags are used", func() {
						It("creates the route", func() {
							session := helpers.CF("create-route", domainName)
							Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session).Should(Say(`Route %s has been created\.`, domainName))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in a hostname", func() {
						It("creates the route with the hostname", func() {
							hostname := "tiramisu"
							session := helpers.CF("create-route", domainName, "-n", hostname)
							Eventually(session).Should(Say(`Creating route %s\.%s for org %s / space %s as %s\.\.\.`, hostname, domainName, orgName, spaceName, userName))
							Eventually(session).Should(Say(`Route %s\.%s has been created\.`, hostname, domainName))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})

		When("the domain does not exist", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("create-route", "some-domain")
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Domain some-domain not found`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the domain is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("create-route")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `DOMAIN` was not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
