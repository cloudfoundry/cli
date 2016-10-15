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
		setAPI()
		loginCF()
		Eventually(CF("delete-org", "-f", orgName), CFLongTimeout).Should(Exit(0))
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				unsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := CF("delete-orphaned-routes", "-f")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				logoutCF()
			})

			It("fails with not logged in message", func() {
				session := CF("delete-orphaned-routes", "-f")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})

		Context("when there no org set", func() {
			BeforeEach(func() {
				logoutCF()
				loginCF()
			})

			It("fails with no targeted org error message", func() {
				session := CF("delete-orphaned-routes", "-f")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No org targeted, use 'cf target -o ORG' to target an org."))
			})
		})

		Context("when there no space set", func() {
			BeforeEach(func() {
				// create a another space, because if the org has only one space it
				// will be automatically targetted
				createSpace(PrefixedRandomName("SPACE"))
				logoutCF()
				loginCF()
				targetOrg(orgName)
			})

			It("fails with no space targeted error message", func() {
				session := CF("delete-orphaned-routes", "-f")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No space targeted, use 'cf target -s SPACE' to target a space"))
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
				WithSimpleApp(func(appDir string) {
					Eventually(CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route"), CFLongTimeout).Should(Exit(0))
				})
				Eventually(CF("apps"), CFLongTimeout).Should(And(Exit(0), Say(fmt.Sprintf("%s\\s+stopped\\s+0/1\\s+%s\\s+%s", appName, DefaultMemoryLimit, DefaultDiskLimit))))

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
