package isolated

import (
	"fmt"
	"regexp"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("remove-network-policy command", func() {
	BeforeEach(func() {
		helpers.SkipIfVersionLessThan(ccversion.MinVersionNetworkingV3)
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("remove-network-policy", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("remove-network-policy - Remove network traffic policy of an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf remove-network-policy SOURCE_APP --destination-app DESTINATION_APP [-s DESTINATION_SPACE_NAME [-o DESTINATION_ORG_NAME]] --protocol (tcp | udp) --port RANGE")))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("   cf remove-network-policy frontend --destination-app backend --protocol tcp --port 8081"))
				Eventually(session).Should(Say("   cf remove-network-policy frontend --destination-app backend -s backend-space -o backend-org --protocol tcp --port 8080-8090"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("   --destination-app      Name of app to connect to"))
				Eventually(session).Should(Say("   --port                 Port or range of ports that destination app is connected with"))
				Eventually(session).Should(Say("   --protocol             Protocol that apps are connected with"))
				Eventually(session).Should(Say(`   -o                     The org of the destination app \(Default: targeted org\)`))
				Eventually(session).Should(Say(`   -s                     The space of the destination app \(Default: targeted space\)`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("   add-network-policy, apps, network-policies"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "remove-network-policy", "some-app", "--destination-app", "some-other-app", "--port", "8080", "--protocol", "tcp")
		})
	})

	When("the org and space are properly targetted", func() {
		var (
			orgName   string
			spaceName string
			appName   string
			appGUID   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.PrefixedRandomName("app")

			helpers.SetupCF(orgName, spaceName)

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
			})

			appGUID = helpers.AppGUID(appName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("an app exists", func() {
			BeforeEach(func() {
				session := helpers.CF("add-network-policy", appName, "--destination-app", appName)

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Adding network policy from app %s to app %s in org %s / space %s as %s\.\.\.`, appName, appName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("network-policies")
				Eventually(session).Should(Say(`Listing network policies in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Consistently(session).ShouldNot(Say("OK"))
				Eventually(session).Should(Say(`source\s+destination\s+protocol\s+ports\s+destination space\s+destination org`))
				Eventually(session).Should(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, appName, appName, spaceName, orgName))
				Eventually(session).Should(Exit(0))
			})

			It("can remove a policy", func() {
				session := helpers.CF("remove-network-policy", appName, "--destination-app", appName, "--port", "8080", "--protocol", "tcp")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Removing network policy from app %s to app %s in org %s / space %s as %s\.\.\.`, appName, appName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("network-policies")
				Eventually(session).Should(Say(`Listing network policies in org %s / space %s as %s\.\.\.`, orgName, spaceName, username))
				Consistently(session).ShouldNot(Say("OK"))
				Eventually(session).Should(Say(`source\s+destination\s+protocol\s+ports\s+destination space\s+destination org`))
				Eventually(session).ShouldNot(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, appName, appName, spaceName, orgName))
				Eventually(session).Should(Exit(0))
			})

			When("an org and space is provided for destination app", func() {
				var (
					sourceOrg     string
					sourceSpace   string
					sourceApp     string
					sourceAppGUID string
				)

				BeforeEach(func() {
					sourceOrg = helpers.NewOrgName()
					sourceSpace = helpers.NewSpaceName()
					sourceApp = helpers.PrefixedRandomName("sourceapp")

					helpers.SetupCF(sourceOrg, sourceSpace)

					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", sourceApp, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
					})

					sourceAppGUID = helpers.AppGUID(sourceApp)

					session := helpers.CF("add-network-policy", sourceApp, "--destination-app", appName, "-o", orgName, "-s", spaceName)
					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(`Adding network policy from app %s in org %s / space %s to app %s in org %s / space %s as %s\.\.\.`, sourceApp, sourceOrg, sourceSpace, appName, orgName, spaceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("network-policies")
					Eventually(session).Should(Say(`Listing network policies in org %s / space %s as %s\.\.\.`, sourceOrg, sourceSpace, username))
					Consistently(session).ShouldNot(Say("OK"))
					Eventually(session).Should(Say(`source\s+destination\s+protocol\s+ports\s+destination space\s+destination org`))
					Eventually(session).Should(Say(`%s\s+%s\s+tcp\s+8080\s+%s\s+%s`, sourceApp, appName, spaceName, orgName))
					Eventually(session).Should(Exit(0))
				})

				It("can remove a policy", func() {
					session := helpers.CF("remove-network-policy", sourceApp, "--destination-app", appName, "-o", orgName, "-s", spaceName, "--port", "8080", "--protocol", "tcp")

					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(`Removing network policy from app %s in org %s / space %s to app %s in org %s / space %s as %s\.\.\.`, sourceApp, sourceOrg, sourceSpace, appName, orgName, spaceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("curl", fmt.Sprintf("/networking/v1/external/policies?id=%s", sourceAppGUID))
					Eventually(session).Should(Exit(0))
					Expect(string(session.Out.Contents())).To(MatchJSON(`{ "total_policies": 0, "policies": [] }`))
				})
			})

			When("the protocol is not provided", func() {
				It("returns a helpful message", func() {
					session := helpers.CF("remove-network-policy", appName, "--destination-app", appName, "--port", "8080")
					Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `--protocol' was not specified"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Say("remove-network-policy - Remove network traffic policy of an app"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the port is not provided", func() {
				It("returns a helpful message", func() {
					session := helpers.CF("remove-network-policy", appName, "--destination-app", appName, "--protocol", "tcp")
					Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `--port' was not specified"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Say("remove-network-policy - Remove network traffic policy of an app"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the policy does not exist", func() {
				It("returns a helpful message and exits 0", func() {
					session := helpers.CF("remove-network-policy", appName, "--destination-app", appName, "--port", "8081", "--protocol", "udp")
					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(`Removing network policy from app %s to app %s in org %s / space %s as %s\.\.\.`, appName, appName, orgName, spaceName, username))
					Eventually(session).Should(Say("Policy does not exist."))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

			})
		})

		When("the source app does not exist", func() {
			It("returns an error", func() {
				session := helpers.CF("remove-network-policy", "pineapple", "--destination-app", appName, "--port", "8080", "--protocol", "tcp")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Removing network policy from app pineapple to app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App pineapple not found"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the dest app does not exist", func() {
			It("returns an error", func() {
				session := helpers.CF("remove-network-policy", appName, "--destination-app", "pineapple", "--port", "8080", "--protocol", "tcp")

				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Removing network policy from app %s to app pineapple in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App pineapple not found"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
