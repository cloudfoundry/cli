package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-route command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("delete-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`delete-route - Delete a route\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf delete-route DOMAIN \[--hostname HOSTNAME\] \[--path PATH\] \[-f\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf delete-route example.com\s+# example.com`))
			Eventually(session).Should(Say(`cf delete-route example.com --hostname myhost\s+# myhost.example.com`))
			Eventually(session).Should(Say(`cf delete-route example.com --hostname myhost --path foo\s+# myhost.example.com/foo`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`-f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname used to identify the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--path\s+Path used to identify the HTTP route`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`delete-orphaned-routes, routes, unmap-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "delete-route", "some-domain")
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
				domainName string
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName()
			})

			When("the route does not exist", func() {
				var (
					domain helpers.Domain
				)

				BeforeEach(func() {
					domain = helpers.NewDomain(orgName, domainName)
					domain.CreatePrivate()
				})

				AfterEach(func() {
					domain.Delete()
				})

				It("warns the user that it has already been deleted and runs to completion without failing", func() {
					session := helpers.CF("delete-route", domainName, "-f")
					Eventually(session).Should(Say(`Deleting route %s\.\.\.`, domainName))
					Eventually(session).Should(Say(`Unable to delete\. Route with domain '%s' not found\.`, domainName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the route exist", func() {
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
						It("deletes the route", func() {
							Eventually(helpers.CF("create-route", domainName)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.\.\.`, domainName))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in a hostname", func() {
						It("deletes the route with the hostname", func() {
							hostname := "tiramisu"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s\.\.\.`, hostname, domainName))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in hostname and path with a leading '/'", func() {
						It("deletes the route with hostname and path", func() {
							hostname := "tiramisu"
							pathString := "/recipes"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "--path", pathString, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s%s\.\.\.`, hostname, domainName, pathString))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in hostname and path without a leading '/'", func() {
						It("deletes the route with hostname and path", func() {
							hostname := "tiramisu"
							pathString := "more-recipes"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "--path", pathString, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s\/%s\.\.\.`, hostname, domainName, pathString))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})

		When("the domain does not exist", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("delete-route", "some-domain", "-f")
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Domain 'some-domain' not found.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the domain is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("delete-route")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `DOMAIN` was not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
