package isolated

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("network-policies command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
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

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "network-policies")
		})

		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with no networking api error message", func() {
				session := helpers.CF("network-policies")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires Network Policy API V1. Your targeted endpoint does not expose it."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the org and space are properly targetted", func() {
		var (
			orgName   string
			spaceName string
			appName   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.PrefixedRandomName("app")

			setupCF(orgName, spaceName)

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
			})

			session := helpers.CF("add-network-policy", appName, "--destination-app", appName)
			Eventually(session).Should(Exit(0))
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when policies exists", func() {
			It("lists all the policies", func() {
				session := helpers.CF("network-policies")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Listing network policies in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Consistently(session).ShouldNot(Say("OK"))
				Eventually(session).Should(Say("source\\s+destination\\s+protocol\\s+ports"))
				Eventually(session).Should(Say("%s\\s+%s\\s+tcp\\s+8080[^-]", appName, appName))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when policies are filtered by a source app", func() {
			var srcAppName string
			BeforeEach(func() {
				srcAppName = helpers.PrefixedRandomName("app")
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", srcAppName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
				})

				session := helpers.CF("add-network-policy", srcAppName, "--destination-app", appName)
				Eventually(session).Should(Exit(0))
			})

			It("lists only policies for which the app is a source", func() {
				session := helpers.CF("network-policies", "--source", srcAppName)

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Listing network policies of app %s in org %s / space %s as %s\.\.\.`, srcAppName, orgName, spaceName, username))
				Eventually(session).Should(Say("source\\s+destination\\s+protocol\\s+ports"))
				Eventually(session).ShouldNot(Say("%s\\s+%s\\s+tcp\\s+8080[^-]", appName, appName))
				Eventually(session).Should(Say("%s\\s+%s\\s+tcp\\s+8080[^-]", srcAppName, appName))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when policies are filtered by a non-existent source app", func() {
			It("returns an error", func() {
				session := helpers.CF("network-policies", "--source", "pineapple")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Listing network policies of app pineapple in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App pineapple not found"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
