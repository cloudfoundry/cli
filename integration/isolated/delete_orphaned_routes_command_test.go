package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-orphaned-routes command", func() {
	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "delete-orphaned-routes")
		})
	})

	Context("when the environment is setup correctly", func() {
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
			domainName = helpers.DomainName()
			appName = helpers.PrefixedRandomName("APP")

			setupCF(orgName, spaceName)
			domain = helpers.NewDomain(orgName, domainName)
			domain.Create()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when there are orphaned routes", func() {
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
					Say("Getting routes as"),
					Say(fmt.Sprintf("Deleting route orphan-1.%s/path-1...", domainName)),
					Say(fmt.Sprintf("Deleting route orphan-2.%s/path-2...", domainName)),
					Say("OK"),
				))
			})
		})

		Context("when there are orphaned routes and bound routes", func() {
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
				Eventually(helpers.CF("apps")).Should(And(Exit(0), Say(fmt.Sprintf("%s\\s+stopped\\s+0/1", appName))))

				boundRoute = helpers.NewRoute(spaceName, domainName, "bound-1", "path-3")
				boundRoute.Create()
				helpers.MapRouteToApplication(appName, boundRoute.Domain, boundRoute.Host, boundRoute.Path)
			})

			It("deletes only the orphaned routes", func() {
				Eventually(helpers.CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(0),
					Say("Getting routes as"),
					Say(fmt.Sprintf("Deleting route orphan-1.%s/path-1...", domainName)),
					Say(fmt.Sprintf("Deleting route orphan-2.%s/path-2...", domainName)),
					Not(Say(fmt.Sprintf("Deleting route bound-1.%s/path-3...", domainName))),
					Say("OK"),
				))
			})
		})

		Context("when there are more than one page of routes", func() {
			BeforeEach(func() {
				var orphanedRoute helpers.Route
				for i := 0; i < 51; i++ {
					orphanedRoute = helpers.NewRoute(spaceName, domainName, fmt.Sprintf("orphan-multi-page-%d", i), "")
					orphanedRoute.Create()
				}
			})
			It("deletes all the orphaned routes", func() {
				session := helpers.CF("delete-orphaned-routes", "-f")

				for i := 0; i < 51; i++ {
					Eventually(session.Out).Should(Say(fmt.Sprintf("Deleting route orphan-multi-page-%d.%s...", i, domainName)))
				}
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the force flag is not given", func() {
			var buffer *Buffer
			BeforeEach(func() {
				orphanedRoute := helpers.NewRoute(spaceName, domainName, "orphan", "path")
				orphanedRoute.Create()
			})

			Context("when the user inputs y", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					buffer.Write([]byte("y\n"))
				})

				It("deletes the orphaned routes", func() {
					session := helpers.CFWithStdin(buffer, "delete-orphaned-routes")
					Eventually(session).Should(Say("Really delete orphaned routes?"))
					Eventually(session).Should(SatisfyAll(
						Exit(0),
						Say("Getting routes as"),
						Say(fmt.Sprintf("Deleting route orphan.%s/path...", domainName)),
						Say("OK"),
					))
				})
			})

			Context("when the user inputs n", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					buffer.Write([]byte("n\n"))
				})

				It("exits without deleting the orphaned routes", func() {
					session := helpers.CFWithStdin(buffer, "delete-orphaned-routes")
					Eventually(session).Should(Say("Really delete orphaned routes?"))
					Eventually(session).Should(SatisfyAll(
						Exit(0),
						Not(Say("Getting routes as")),
						Not(Say(fmt.Sprintf("Deleting route orphan.%s/path...", domainName))),
						Not(Say("OK")),
					))
				})
			})
		})

		Context("when there are no orphaned routes", func() {
			var (
				boundRoute helpers.Route
			)

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})
				Eventually(helpers.CF("apps")).Should(And(Exit(0), Say(fmt.Sprintf("%s\\s+stopped\\s+0/1", appName))))

				boundRoute = helpers.NewRoute(spaceName, domainName, "bound-route", "bound-path")
				boundRoute.Create()
				helpers.MapRouteToApplication(appName, boundRoute.Domain, boundRoute.Host, boundRoute.Path)
			})

			It("displays OK without deleting any routes", func() {
				Eventually(helpers.CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(0),
					Say("Getting routes as"),
					Not(Say(fmt.Sprintf("Deleting route bound-route.%s/bound-path...", domainName))),
					Say("OK"),
				))
			})
		})

		Context("when the orphaned routes are attached to both shared and private domains", func() {
			var (
				orphanedRoute1   helpers.Route
				orphanedRoute2   helpers.Route
				sharedDomainName string
			)

			BeforeEach(func() {
				sharedDomainName = helpers.DomainName()
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
					Say("Getting routes as"),
					Say(fmt.Sprintf("Deleting route orphan-1.%s/path-1...", domainName)),
					Say(fmt.Sprintf("Deleting route orphan-2.%s/path-2...", sharedDomainName)),
					Say("OK"),
				))
			})
		})
	})
})
