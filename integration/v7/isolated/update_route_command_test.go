package isolated

import (
	. "code.cloudfoundry.org/cli/v9/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/v9/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-route command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("update-route", "ROUTES", "Update a route by route specific options, e.g. load balancing algorithm"))
		})

		It("displays the help information", func() {
			session := helpers.CF("update-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`update-route - Update a route by route specific options, e.g. load balancing algorithm\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`Update an existing HTTP route:\n`))
			Eventually(session).Should(Say(`cf update-route DOMAIN \[--hostname HOSTNAME\] \[--path PATH\] \[--option OPTION=VALUE\] \[--remove-option OPTION\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf update-route example.com -o loadbalancing=round-robin`))
			Eventually(session).Should(Say(`cf update-route example.com -o loadbalancing=least-connection`))
			Eventually(session).Should(Say(`cf update-route example.com -r loadbalancing`))
			Eventually(session).Should(Say(`cf update-route example.com --hostname myhost --path foo -o loadbalancing=round-robin`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--path\s+Path for the HTTP route`))
			Eventually(session).Should(Say(`--option, -o\s+Set the value of a per-route option`))
			Eventually(session).Should(Say(`--remove-option, -r\s+Remove an option with the given name`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`check-route, domains, map-route, routes, unmap-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "update-route", "some-domain")
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
					option   string
					path     string
				)

				BeforeEach(func() {
					domain = helpers.NewDomain(orgName, domainName)
					hostname = "key-lime-pie"
					path = "/a"
					domain.CreatePrivate()
					Eventually(helpers.CF("create-route", domainName, "--hostname", hostname, "--path", path)).Should(Exit(0))
				})

				AfterEach(func() {
					domain.Delete()
				})
				When("a route option is specified", func() {
					It("updates the route and runs to completion without failing", func() {
						option = "loadbalancing=round-robin"
						session := helpers.CF("update-route", domainName, "--hostname", hostname, "--path", path, "--option", option)
						Eventually(session).Should(Say(`Updating route %s\.%s%s for org %s / space %s as %s\.\.\.`, hostname, domainName, path, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Route %s\.%s%s has been updated`, hostname, domainName, path))
						Eventually(session).Should(Say(`OK`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("route options are not specified", func() {
					It("gives an error message and fails", func() {
						session := helpers.CF("update-route", domainName, "--hostname", hostname, "--path", path)
						Eventually(session.Err).Should(Say(`Route option support: 'No options were specified for the update of the Route %s\.%s\%s`, hostname, domainName, path))
						Eventually(session).Should(Exit(1))
					})
				})

				When("route options are specified in the wrong format", func() {
					It("gives an error message and fails", func() {
						session := helpers.CF("update-route", domainName, "--hostname", hostname, "--path", path, "--option", "loadbalancing")
						Eventually(session.Err).Should(Say(`Route option '%s' for route with host '%s', domain '%s', and path '%s' was specified incorrectly. Please use key-value pair format key=value.`, "loadbalancing", hostname, domainName, path))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the route does not exist", func() {
				var (
					domain   helpers.Domain
					hostname string
					option   string
				)

				BeforeEach(func() {
					domain = helpers.NewDomain(orgName, domainName)
					hostname = "key-lime-pie"
					option = "loadbalancing=round-robin"
					domain.CreatePrivate()
				})

				AfterEach(func() {
					domain.Delete()
				})

				It("gives an error message", func() {
					session := helpers.CF("update-route", domainName, "--hostname", hostname, "--option", option)
					Eventually(session).Should(Say(`Updating route %s\.%s for org %s / space %s as %s\.\.\.`, hostname, domainName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say(`API endpoint not found at`))
					Eventually(session).Should(Exit(1))
				})
			})

		})

		When("the domain does not exist", func() {
			It("gives an error message and exits", func() {
				session := helpers.CF("update-route", "some-domain")
				Eventually(session.Err).Should(Say(`Domain '%s' not found.`, "some-domain"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the domain is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("update-route")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `DOMAIN` was not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
