package readonly

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("apps command", func() {
	var (
		orgName   string
		spaceName string
		appName1  string
		appName2  string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName1 = helpers.PrefixedRandomName("app1")
		appName2 = helpers.PrefixedRandomName("app2")
		helpers.TurnOffExperimental()
	})

	AfterEach(func() {
		helpers.TurnOnExperimental()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("apps", "APPS", "List all apps in the target space"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("apps", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("apps - List all apps in the target space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf apps"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("a"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("events, logs, map-route, push, restart, scale, start, stop"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "apps")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("with no apps", func() {
			It("displays empty list", func() {
				session := helpers.CF("apps")
				Eventually(session).Should(Say(`Getting apps in org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Eventually(session).Should(Say("No apps found"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("with existing apps", func() {
			var domainName string

			BeforeEach(func() {
				domainName = helpers.DefaultSharedDomain()
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName2)).Should(Exit(0))
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName1)).Should(Exit(0))
				})

			})

			It("displays apps in the list", func() {
				session := helpers.CF("apps")
				Eventually(session).Should(Say(`Getting apps in org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Eventually(session).Should(Say(`name\s+requested state\s+processes\s+routes`))
				Eventually(session).Should(Say(`%s\s+started\s+web:1/1, console:0/0\s+%s\.%s`, appName1, appName1, domainName))
				Eventually(session).Should(Say(`%s\s+started\s+web:1/1, console:0/0\s+%s\.%s`, appName2, appName2, domainName))

				Eventually(session).Should(Exit(0))
			})

			When("one app is stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("stop", appName1)).Should(Exit(0))
				})

				It("displays app as stopped", func() {
					session := helpers.CF("apps")
					Eventually(session).Should(Say(`Getting apps in org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
					Eventually(session).Should(Say(`name\s+requested state\s+processes\s+routes`))
					Eventually(session).Should(Say(`%s\s+stopped\s+web:0/1, console:0/0\s+%s\.%s`, appName1, appName1, domainName))
					Eventually(session).Should(Say(`%s\s+started\s+web:1/1, console:0/0\s+%s\.%s`, appName2, appName2, domainName))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the --labels flag is given", func() {

				BeforeEach(func() {
					Eventually(helpers.CF("set-label", "app", appName1, "environment=production", "tier=backend")).Should(Exit(0))
					Eventually(helpers.CF("set-label", "app", appName2, "environment=staging", "tier=frontend")).Should(Exit(0))
				})

				It("displays apps with provided labels", func() {
					session := helpers.CF("apps", "--labels", "environment in (production,staging),tier in (backend)")
					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say(appName1))
					Expect(session).ShouldNot(Say(appName2))
				})
			})
		})

	})
})
