package integration

import (
	"fmt"

	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-orphaned-routes command", func() {
	var (
		orgName    string
		spaceName  string
		domainName string
		appName    string
		domain     Domain
	)

	BeforeEach(func() {
		Skip("until #131127157")
		orgName = PrefixedRandomName("ORG")
		spaceName = PrefixedRandomName("SPACE")
		domainName = fmt.Sprintf("%s.com", PrefixedRandomName("DOMAIN"))
		appName = PrefixedRandomName("APP")

		setupCF(orgName, spaceName)
		domain = NewDomain(orgName, domainName)
		domain.Create()
	})

	AfterEach(func() {
		Eventually(CF("delete-org", "-f", orgName), CFLongTimeout).Should(Exit(0))
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				unsetAPI()
			})

			AfterEach(func() {
				setAPI()
				loginCF()
			})

			It("fails with no API endpoint set message", func() {
				Eventually(CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(1),
					Say("FAILED\nNo API endpoint set. Use 'cf login' or 'cf api' to target an endpoint.")),
				)
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				logoutCF()
			})

			AfterEach(func() {
				loginCF()
			})

			It("fails with not logged in message", func() {
				Eventually(CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(1),
					Say("FAILED\nNot logged in. Use 'cf login' to log in.")),
				)
			})
		})

		Context("when there no space set", func() {
			BeforeEach(func() {
				logoutCF()
				loginCF()
			})

			It("fails with no targeted space error message", func() {
				Eventually(CF("delete-orphaned-routes", "-f")).Should(SatisfyAll(
					Exit(1),
					Say("FAILED\nFailed fetching routes.\nServer error, status code: 404, error code: 40004, message: The app space could not be found: routes")),
				)
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		var (
			orphanedRoute1 Route
			orphanedRoute2 Route
		)

		BeforeEach(func() {
			orphanedRoute1 = NewRoute(spaceName, domainName, "orphan-1", "path-1")
			orphanedRoute2 = NewRoute(spaceName, domainName, "orphan-2", "path-2")
			orphanedRoute1.Create()
			orphanedRoute2.Create()
		})

		AfterEach(func() {
			orphanedRoute1.Delete()
			orphanedRoute2.Delete()
		})

		Context("when there are orphaned routes", func() {
			It("deletes all the orphaned routes", func() {
				Eventually(CF("delete-orphaned-routes", "-f"), CFLongTimeout).Should(SatisfyAll(
					Exit(0),
					Say("Getting routes as"),
					Say(fmt.Sprintf("Deleting route orphan-1.%s/path-1...", domainName)),
					Say(fmt.Sprintf("Deleting route orphan-2.%s/path-2...", domainName)),
					Say("OK"),
				))
			})
		})

		Context("when there are orphaned routes and bound routes", func() {
			var boundRoute Route

			BeforeEach(func() {
				Eventually(CF("push", appName, "-p", "./assets/dora", "-m", DefaultMemoryLimit, "-k", DefaultDiskLimit, "--no-route"), CFLongTimeout).Should(Exit(0))
				Eventually(CF("apps"), CFLongTimeout).Should(And(Exit(0), Say(fmt.Sprintf("%s\\s+started\\s+1/1\\s+%s\\s+%s", appName, DefaultMemoryLimit, DefaultDiskLimit))))

				boundRoute = NewRoute(spaceName, domainName, "bound-1", "path-3")
				boundRoute.Create()
				BindRouteToApplication(appName, boundRoute.Domain, boundRoute.Host, boundRoute.Path)
			})

			AfterEach(func() {
				UnbindRouteToApplication(appName, boundRoute.Domain, boundRoute.Host, boundRoute.Path)
				boundRoute.Delete()
			})

			It("deletes only the orphaned routes", func() {
				Eventually(CF("delete-orphaned-routes", "-f"), CFLongTimeout).Should(SatisfyAll(
					Exit(0),
					Say("Getting routes as"),
					Say(fmt.Sprintf("Deleting route orphan-1.%s/path-1...", domainName)),
					Say(fmt.Sprintf("Deleting route orphan-2.%s/path-2...", domainName)),
					Not(Say(fmt.Sprintf("Deleting route bound-1.%s/path-3...", domainName))),
					Say("OK"),
				))
			})
		})
	})
})
