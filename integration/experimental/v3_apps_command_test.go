package experimental

import (
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
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-apps", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-apps - List all apps in the target space"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-apps"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-apps")
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithV3Version("3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-apps")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no space targeted error message", func() {
				session := helpers.CF("v3-apps")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			setupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("with no apps", func() {
			It("displays empty list", func() {
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say("Getting apps in org %s / space %s as %s\\.\\.\\.", orgName, spaceName, userName))
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

				domainName = defaultSharedDomain()
			})

			It("displays apps in the list", func() {
				session := helpers.CF("v3-apps")
				Eventually(session).Should(Say("Getting apps in org %s / space %s as %s\\.\\.\\.", orgName, spaceName, userName))
				Eventually(session).Should(Say("name\\s+requested state\\s+processes\\s+routes"))
				Eventually(session).Should(Say("%s\\s+started\\s+web:1/1, console:0/0, rake:0/0\\s+%s\\.%s", appName1, appName1, domainName))
				Eventually(session).Should(Say("%s\\s+started\\s+web:1/1, console:0/0, rake:0/0\\s+%s\\.%s", appName2, appName2, domainName))

				Eventually(session).Should(Exit(0))
			})

			Context("when one app is stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("stop", appName1)).Should(Exit(0))
				})

				It("displays app as stopped", func() {
					session := helpers.CF("v3-apps")
					Eventually(session).Should(Say("Getting apps in org %s / space %s as %s\\.\\.\\.", orgName, spaceName, userName))
					Eventually(session).Should(Say("name\\s+requested state\\s+processes\\s+routes"))
					Eventually(session).Should(Say("%s\\s+stopped\\s+web:0/1, console:0/0, rake:0/0\\s+%s\\.%s", appName1, appName1, domainName))
					Eventually(session).Should(Say("%s\\s+started\\s+web:1/1, console:0/0, rake:0/0\\s+%s\\.%s", appName2, appName2, domainName))

					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
