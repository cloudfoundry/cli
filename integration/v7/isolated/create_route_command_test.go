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

var _ = Describe("create-route command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("create-route", "ROUTES", "Create a route for later use"))
		})

		It("displays the help information", func() {
			session := helpers.CF("create-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`create-route - Create a route for later use\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`Create an HTTP route:\n`))
			Eventually(session).Should(Say(`cf create-route DOMAIN \[--hostname HOSTNAME\] \[--path PATH\]\n`))
			Eventually(session).Should(Say(`Create a TCP route:\n`))
			Eventually(session).Should(Say(`cf create-route DOMAIN \[--port PORT\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf create-route example.com\s+# example.com`))
			Eventually(session).Should(Say(`cf create-route example.com --hostname myapp\s+# myapp.example.com`))
			Eventually(session).Should(Say(`cf create-route example.com --hostname myapp --path foo\s+# myapp.example.com/foo`))
			Eventually(session).Should(Say(`cf create-route example.com --port 5000\s+# example.com:5000`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--path\s+Path for the HTTP route`))
			Eventually(session).Should(Say(`--port\s+Port for the TCP route \(default: random port\)`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`check-route, domains, map-route, routes, unmap-route`))

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
					Eventually(session.Err).Should(Say(`Route already exists with host '%s' for domain '%s'\.`, hostname, domainName))
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

					When("passing in hostname and path with a leading '/'", func() {
						It("creates the route with hostname and path", func() {
							hostname := "tiramisu"
							pathString := "/recipes"
							session := helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)
							Eventually(session).Should(Say(`Creating route %s\.%s%s for org %s / space %s as %s\.\.\.`, hostname, domainName, pathString, orgName, spaceName, userName))
							Eventually(session).Should(Say(`Route %s\.%s%s has been created\.`, hostname, domainName, pathString))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in hostname and path without a leading '/'", func() {
						It("creates the route with hostname and path", func() {
							hostname := "tiramisu"
							pathString := "more-recipes"
							session := helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)
							Eventually(session).Should(Say(`Creating route %s\.%s\/%s for org %s / space %s as %s\.\.\.`, hostname, domainName, pathString, orgName, spaceName, userName))
							Eventually(session).Should(Say(`Route %s\.%s\/%s has been created\.`, hostname, domainName, pathString))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the domain is a shared HTTP domain", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain("", domainName)
						domain.CreateShared()
					})

					AfterEach(func() {
						domain.DeleteShared()
					})

					When("no flags are used", func() {
						It("errors indicating that hostname is missing", func() {
							session := helpers.CF("create-route", domainName)
							Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session.Out).Should(Say(`FAILED`))
							Eventually(session.Err).Should(Say(`Missing host. Routes in shared domains must have a host defined.`))
							Eventually(session).Should(Exit(1))
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

					When("passing in a hostname and path", func() {
						It("creates the route with the hostname", func() {
							hostname := "tiramisu"
							path := "lion"
							session := helpers.CF("create-route", domainName, "-n", hostname, "--path", path)
							Eventually(session).Should(Say(`Creating route %s\.%s\/%s for org %s / space %s as %s\.\.\.`, hostname, domainName, path, orgName, spaceName, userName))
							Eventually(session).Should(Say(`Route %s\.%s\/%s has been created\.`, hostname, domainName, path))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the domain is a shared TCP domain", func() {
					var (
						domain      helpers.Domain
						routerGroup helpers.RouterGroup
					)

					BeforeEach(func() {
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

					When("passing in a port", func() {
						It("creates the route with the port", func() {
							port := 1029
							session := helpers.CF("create-route", domainName, "--port", fmt.Sprintf("%d", port))
							Eventually(session).Should(Say(`Creating route %s:%d for org %s / space %s as %s\.\.\.`, domainName, port, orgName, spaceName, userName))
							Eventually(session).Should(Say(`Route %s:%d has been created\.`, domainName, port))
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
				Eventually(session.Err).Should(Say(`Domain 'some-domain' not found.`))
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
