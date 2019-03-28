package isolated

import (
	"fmt"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-orphaned-routes command", func() {
	Describe("help text", func() {
		It("displays the help information", func() {
			session := helpers.CF("delete-orphaned-routes", "--help")
			Eventually(session).Should(Say("NAME:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("delete-orphaned-routes - Delete all orphaned routes in the currently targeted space (i.e. those that are not mapped to an app)\n")))
			Eventually(session).Should(Say("USAGE:\n"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf delete-orphaned-routes [-f]")))
			Eventually(session).Should(Say("OPTIONS:\n"))
			Eventually(session).Should(Say(`-f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say("SEE ALSO:\n"))
			Eventually(session).Should(Say("delete-route, routes"))
			Eventually(session).Should(Exit(0))
		})
	})
	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "delete-orphaned-routes")
		})
	})

	When("the environment is setup correctly", func() {
		var (
			orgName    string
			spaceName  string
			domainName string
			appName    string
			domain     helpers.Domain
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			domainName = helpers.NewDomainName()
			appName = helpers.PrefixedRandomName("APP")

			helpers.SetupCF(orgName, spaceName)
			domain = helpers.NewDomain(orgName, domainName)
			domain.Create()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("there are orphaned routes", func() {
			var (
				orphanedRoute1 helpers.Route
				orphanedRoute2 helpers.Route
			)

			BeforeEach(func() {
				orphanedRoute1 = helpers.NewRoute(spaceName, domainName, "orphan-1", "path-1")
				orphanedRoute2 = helpers.NewRoute(spaceName, domainName, "orphan-2", "path-2")
				orphanedRoute1.Create()
				orphanedRoute2.Create()
			})

			It("deletes all the orphaned routes", func() {
				Eventually(helpers.CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(0),
					Say("Deleting routes as"),
					Say("OK"),
				))

				Eventually(helpers.CF("routes")).Should(Say("No routes found"))
			})
		})

		When("there are orphaned routes and bound routes", func() {
			var (
				orphanedRoute1 helpers.Route
				orphanedRoute2 helpers.Route
				boundRoute     helpers.Route
			)

			BeforeEach(func() {
				orphanedRoute1 = helpers.NewRoute(spaceName, domainName, "orphan-1", "path-1")
				orphanedRoute2 = helpers.NewRoute(spaceName, domainName, "orphan-2", "path-2")
				orphanedRoute1.Create()
				orphanedRoute2.Create()

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				boundRoute = helpers.NewRoute(spaceName, domainName, "bound-1", "path-3")
				boundRoute.Create()
				helpers.MapRouteToApplication(appName, boundRoute.Domain, boundRoute.Host, boundRoute.Path)
			})

			It("deletes only the orphaned routes", func() {
				Eventually(helpers.CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(0),
					Say("Deleting routes as"),
					Say("OK"),
				))

				Eventually(helpers.CF("routes")).Should(SatisfyAll(
					Say("bound-1.*path-3"),
					Not(Say("orphan-1.*path-1")),
					Not(Say("orphan-2.*path-2")),
				))
			})
		})

		When("there are more than one page of routes", func() {
			BeforeEach(func() {
				var orphanedRoute helpers.Route
				for i := 0; i < 51; i++ {
					orphanedRoute = helpers.NewRoute(spaceName, domainName, fmt.Sprintf("orphan-multi-page-%d", i), "")
					orphanedRoute.Create()
				}
			})
			It("deletes all the orphaned routes", func() {
				session := helpers.CF("delete-orphaned-routes", "-f")

				Eventually(session).Should(SatisfyAll(
					Exit(0),
					Say("OK"),
				))

				Eventually(helpers.CF("routes")).Should(Say("No routes found"))
			})
		})

		When("the force flag is not given", func() {
			var buffer *Buffer
			BeforeEach(func() {
				orphanedRoute := helpers.NewRoute(spaceName, domainName, "orphan", "path")
				orphanedRoute.Create()
			})

			When("the user inputs y", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					_, err := buffer.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("deletes the orphaned routes", func() {
					session := helpers.CFWithStdin(buffer, "delete-orphaned-routes")
					Eventually(session).Should(Say("Really delete orphaned routes?"))
					Eventually(session).Should(SatisfyAll(
						Exit(0),
						Say("OK"),
					))
				})
			})

			When("the user inputs n", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					_, err := buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("exits without deleting the orphaned routes", func() {
					session := helpers.CFWithStdin(buffer, "delete-orphaned-routes")
					Eventually(session).Should(Say("Really delete orphaned routes?"))
					Eventually(session).Should(SatisfyAll(
						Exit(0),
						Not(Say("OK")),
					))
				})
			})
		})

		When("there are no orphaned routes", func() {
			var (
				boundRoute helpers.Route
			)

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				boundRoute = helpers.NewRoute(spaceName, domainName, "bound-route", "bound-path")
				boundRoute.Create()
				helpers.MapRouteToApplication(appName, boundRoute.Domain, boundRoute.Host, boundRoute.Path)
			})

			It("displays OK without deleting any routes", func() {
				Eventually(helpers.CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(0),
					Say("OK"),
				))
			})
		})

		When("the orphaned routes are attached to both shared and private domains", func() {
			var (
				orphanedRoute1   helpers.Route
				orphanedRoute2   helpers.Route
				sharedDomainName string
			)

			BeforeEach(func() {
				sharedDomainName = helpers.NewDomainName()
				sharedDomain := helpers.NewDomain(orgName, sharedDomainName)
				sharedDomain.Create()
				sharedDomain.Share()

				orphanedRoute1 = helpers.NewRoute(spaceName, domainName, "orphan-1", "path-1")
				orphanedRoute2 = helpers.NewRoute(spaceName, sharedDomainName, "orphan-2", "path-2")
				orphanedRoute1.Create()
				orphanedRoute2.Create()
			})

			It("deletes both the routes", func() {
				Eventually(helpers.CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(0),
					Say("OK"),
				))
			})
		})
	})
})
