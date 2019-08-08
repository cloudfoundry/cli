package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("check-route command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("check-route", "ROUTES", "Perform a check to determine whether a route currently exists or not"))
		})

		It("displays the help information", func() {
			session := helpers.CF("check-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`check-route - Perform a check to determine whether a route currently exists or not\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf check-route DOMAIN \[--hostname HOSTNAME\] \[--path PATH\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf check-route example.com                      # example.com`))
			Eventually(session).Should(Say(`cf check-route example.com -n myhost --path foo # myhost.example.com/foo`))
			Eventually(session).Should(Say(`cf check-route example.com --path foo           # example.com/foo`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname used to identify the HTTP route`))
			Eventually(session).Should(Say(`--path\s+Path for the route`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`create-route, delete-route, routes`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "check-route", "some-domain")
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

				It("tells the user the route exists and exits without failing", func() {
					session := helpers.CF("check-route", domainName, "--hostname", hostname)
					Eventually(session).Should(Say(`Checking for route\.\.\.`))
					Eventually(session).Should(Say(`Route '%s\.%s' does exist\.`, hostname, domainName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the route does not already exist", func() {
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
						session := helpers.CF("check-route", domainName)
						Eventually(session).Should(Say(`Checking for route\.\.\.`))
						Eventually(session).Should(Say(`Route '%s' does not exist\.`, domainName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("passing in a hostname", func() {
					It("checks the route with the hostname", func() {
						hostname := "tiramisu"
						session := helpers.CF("check-route", domainName, "-n", hostname)
						Eventually(session).Should(Say(`Checking for route\.\.\.`))
						Eventually(session).Should(Say(`Route '%s\.%s' does not exist\.`, hostname, domainName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("passing in hostname and path with a leading '/'", func() {
					It("checks the route with hostname and path", func() {
						hostname := "tiramisu"
						pathString := "/recipes"
						session := helpers.CF("check-route", domainName, "-n", hostname, "--path", pathString)
						Eventually(session).Should(Say(`Checking for route\.\.\.`))
						Eventually(session).Should(Say(`Route '%s\.%s%s' does not exist\.`, hostname, domainName, pathString))
						Eventually(session).Should(Exit(0))
					})
				})

				When("passing in hostname and path without a leading '/'", func() {
					It("checks the route with hostname and path", func() {
						hostname := "tiramisu"
						pathString := "more-recipes"
						session := helpers.CF("check-route", domainName, "-n", hostname, "--path", pathString)
						Eventually(session).Should(Say(`Checking for route\.\.\.`))
						Eventually(session).Should(Say(`Route '%s\.%s\/%s' does not exist\.`, hostname, domainName, pathString))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		When("the domain does not exist", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("check-route", "some-domain")
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Domain 'some-domain' not found.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the domain is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("check-route")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `DOMAIN` was not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
