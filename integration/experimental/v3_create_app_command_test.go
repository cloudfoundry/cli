package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-create-app command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-create-app", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-create-app - Create a V3 App"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-create-app APP_NAME \\[--app-type \\(buildpack | docker\\)\\]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--app-type\\s+App lifecycle type to stage and run the app \\(Default: buildpack\\)"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-create-app")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-create-app", appName)
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when app type is not supported", func() {
		It("tells the user that the app type is incorrect, prints help text, and exits 1", func() {
			session := helpers.CF("v3-create-app", appName, "--app-type", "unknown-app-type")

			Eventually(session.Err).Should(Say("Incorrect Usage: Invalid value `unknown-app-type' for option `--app-type'. Allowed values are: buildpack or docker"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-create-app", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
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
				session := helpers.CF("v3-create-app", appName)
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
				session := helpers.CF("v3-create-app", appName)
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
				session := helpers.CF("v3-create-app", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("v3-create-app", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error message", func() {
				session := helpers.CF("v3-create-app", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app does not exist", func() {
			It("creates the app with default app type buildpack", func() {
				session := helpers.CF("v3-create-app", appName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating V3 app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("v3-app", appName)
				Eventually(session).Should(Say("buildpacks:"))
				Eventually(session).Should(Exit(0))
			})

			Context("when app type is specified", func() {
				It("creates the app with the specified app type", func() {
					session := helpers.CF("v3-create-app", appName, "--app-type", "docker")
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating V3 app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("v3-app", appName)
					Eventually(session).Should(Say("docker image:"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the app already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("v3-create-app", appName)).Should(Exit(0))
			})

			It("fails to create the app", func() {
				session := helpers.CF("v3-create-app", appName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating V3 app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s already exists", appName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
