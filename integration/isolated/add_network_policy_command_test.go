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

var _ = Describe("add-network-policy command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("add-network-policy", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("add-network-policy - Create policy to allow direct network traffic from one app to another"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf add-network-policy SOURCE_APP --destination-app DESTINATION_APP [(--protocol (tcp | udp) --port RANGE)]")))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("   cf add-network-policy frontend --destination-app backend --protocol tcp --port 8081"))
				Eventually(session).Should(Say("   cf add-network-policy frontend --destination-app backend --protocol tcp --port 8080-8090"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("   --destination-app      Name of app to connect to"))
				Eventually(session).Should(Say("   --port                 Port or range of ports for connection to destination app \\(Default: 8080\\)"))
				Eventually(session).Should(Say("   --protocol             Protocol to connect apps with \\(Default: tcp\\)"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("   apps, network-policies"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "add-network-policy", "some-app", "--destination-app", "some-other-app")
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
				session := helpers.CF("add-network-policy", "some-app", "--destination-app", "some-other-app")
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
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when an app exists", func() {
			It("creates a policy", func() {
				session := helpers.CF("add-network-policy", appName, "--destination-app", appName, "--port", "8080-8090", "--protocol", "udp")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Adding network policy to app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("network-policies")
				Eventually(session).Should(Say(`Listing network policies in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Consistently(session).ShouldNot(Say("OK"))
				Eventually(session).Should(Say("source\\s+destination\\s+protocol\\s+ports"))
				Eventually(session).Should(Say("%s\\s+%s\\s+udp\\s+8080-8090", appName, appName))
				Eventually(session).Should(Exit(0))
			})

			Context("when port and protocol are not specified", func() {
				It("creates a policy with the default values", func() {
					session := helpers.CF("add-network-policy", appName, "--destination-app", appName)

					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(`Adding network policy to app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("network-policies")
					Eventually(session).Should(Say(`Listing network policies in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
					Consistently(session).ShouldNot(Say("OK"))
					Eventually(session).Should(Say("source\\s+destination\\s+protocol\\s+ports"))
					Eventually(session).Should(Say("%s\\s+%s\\s+tcp\\s+8080[^-]", appName, appName))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the source app does not exist", func() {
			It("returns an error", func() {
				session := helpers.CF("add-network-policy", "pineapple", "--destination-app", appName)

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Adding network policy to app pineapple in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App pineapple not found"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the dest app does not exist", func() {
			It("returns an error", func() {
				session := helpers.CF("add-network-policy", appName, "--destination-app", "pineapple")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Adding network policy to app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App pineapple not found"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when port is specified but protocol is not", func() {
			It("returns an error", func() {
				session := helpers.CF("add-network-policy", appName, "--destination-app", appName, "--port", "8080")

				Eventually(session.Err).Should(Say("Incorrect Usage: --protocol and --port flags must be specified together"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when protocol is specified but port is not", func() {
			It("returns an error", func() {
				session := helpers.CF("add-network-policy", appName, "--destination-app", appName, "--protocol", "tcp")

				Eventually(session.Err).Should(Say("Incorrect Usage: --protocol and --port flags must be specified together"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
