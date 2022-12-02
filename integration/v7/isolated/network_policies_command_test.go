package isolated

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("network-policies command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("network-policies", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("network-policies - List direct network traffic policies"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf network-policies [--source SOURCE_APP]")))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("   --source      Source app to filter results by"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("   add-network-policy, apps, remove-network-policy"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "network-policies")
		})
	})

	When("the org and space are properly targeted", func() {
		var (
			orgName   string
			spaceName string
			appName   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.PrefixedRandomName("app")

			helpers.SetupCF(orgName, spaceName)

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
			})

			session := helpers.CF("add-network-policy", appName, appName)
			Eventually(session).Should(Exit(0))
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("policies exists", func() {
			It("lists all the policies", func() {
				session := helpers.CF("network-policies")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Listing network policies in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Consistently(session).ShouldNot(Say("OK"))
				Eventually(session).Should(Say(`source\s+destination\s+protocol\s+ports\s+destination space\s+destination org`))
				Eventually(session).Should(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, appName, appName, spaceName, orgName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("policy has a destination in another org and space", func() {
			var (
				destOrg   string
				destSpace string
				destApp   string
			)

			BeforeEach(func() {
				destOrg = helpers.NewOrgName()
				destSpace = helpers.NewSpaceName()
				destApp = helpers.PrefixedRandomName("destapp")

				helpers.SetupCF(destOrg, destSpace)

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", destApp, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
				})

				helpers.SetupCF(orgName, spaceName)

				session := helpers.CF("add-network-policy", appName, destApp, "-o", destOrg, "-s", destSpace)
				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Adding network policy from app %s in org %s / space %s to app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, destApp, destOrg, destSpace, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			It("lists the policy", func() {
				session := helpers.CF("network-policies")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Listing network policies in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Eventually(session).Should(Say(`source\s+destination\s+protocol\s+ports\s+destination space\s+destination org`))
				Eventually(session).Should(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, appName, destApp, destSpace, destOrg))
				Eventually(session).Should(Exit(0))
			})
		})

		When("policies are filtered by a source app", func() {
			var srcAppName string
			BeforeEach(func() {
				srcAppName = helpers.PrefixedRandomName("app")
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", srcAppName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
				})

				session := helpers.CF("add-network-policy", srcAppName, appName)
				Eventually(session).Should(Exit(0))
			})

			It("lists only policies for which the app is a source", func() {
				session := helpers.CF("network-policies", "--source", srcAppName)

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Listing network policies of app %s in org %s / space %s as %s\.\.\.`, srcAppName, orgName, spaceName, username))
				Eventually(session).Should(Say(`source\s+destination\s+protocol\s+ports\s+destination space\s+destination org`))
				Eventually(session).ShouldNot(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, appName, appName, spaceName, orgName))
				Eventually(session).Should(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, srcAppName, appName, spaceName, orgName))
				Eventually(session).Should(Exit(0))
			})

			When("policy has a destination in another space", func() {
				var (
					destOrg   string
					destSpace string
					destApp   string
				)

				BeforeEach(func() {
					destOrg = helpers.NewOrgName()
					destSpace = helpers.NewSpaceName()
					destApp = helpers.PrefixedRandomName("destapp")

					helpers.SetupCF(destOrg, destSpace)

					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", destApp, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
					})

					helpers.SetupCF(orgName, spaceName)

					session := helpers.CF("add-network-policy", appName, destApp, "-o", destOrg, "-s", destSpace)
					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(`Adding network policy from app %s in org %s / space %s to app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, destApp, destOrg, destSpace, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

				It("lists only policies for which the app is a source", func() {
					session := helpers.CF("network-policies", "--source", appName)

					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(`Listing network policies of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
					Eventually(session).Should(Say(`source\s+destination\s+protocol\s+ports\s+destination space\s+destination org`))
					Eventually(session).Should(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, appName, appName, spaceName, orgName))
					Eventually(session).Should(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, appName, destApp, destSpace, destOrg))
					Eventually(session).ShouldNot(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, srcAppName, appName, spaceName, orgName))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("policies are filtered by a nonexistent source app", func() {
			It("returns an error", func() {
				session := helpers.CF("network-policies", "--source", "pineapple")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Listing network policies of app pineapple in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App 'pineapple' not found"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
