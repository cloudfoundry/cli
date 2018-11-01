// +build !partialPush

package experimental

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-apps command", func() {
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
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-apps", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-apps - List all apps in the target space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-apps"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-apps")
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		When("the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithAPIVersions(helpers.DefaultV2Version, ccversion.MinV3ClientVersion)
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`This command requires CF API version 3\.27\.0 or higher\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-apps")
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
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say(`Getting apps in org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Eventually(session).Should(Say("No apps found"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("with existing apps", func() {
			var domainName string

			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName2)).Should(Exit(0))
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName1)).Should(Exit(0))
				})

				domainName = helpers.DefaultSharedDomain()
			})

			It("displays apps in the list", func() {
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say(`Getting apps in org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
				Eventually(session).Should(Say(`name\s+requested state\s+processes\s+routes`))
				Eventually(session).Should(Say(`%s\s+started\s+web:1/1, console:0/0, rake:0/0\s+%s\.%s`, appName1, appName1, domainName))
				Eventually(session).Should(Say(`%s\s+started\s+web:1/1, console:0/0, rake:0/0\s+%s\.%s`, appName2, appName2, domainName))

				Eventually(session).Should(Exit(0))
			})

			When("one app is stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("stop", appName1)).Should(Exit(0))
				})

				It("displays app as stopped", func() {
					session := helpers.CF("v3-apps")
					Eventually(session).Should(Say(`Getting apps in org %s / space %s as %s\.\.\.`, orgName, spaceName, userName))
					Eventually(session).Should(Say(`name\s+requested state\s+processes\s+routes`))
					Eventually(session).Should(Say(`%s\s+stopped\s+web:0/1, console:0/0, rake:0/0\s+%s\.%s`, appName1, appName1, domainName))
					Eventually(session).Should(Say(`%s\s+started\s+web:1/1, console:0/0, rake:0/0\s+%s\.%s`, appName2, appName2, domainName))

					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
