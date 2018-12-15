// +build !partialPush

package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-cancel-zdt-push command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.NewAppName()
		helpers.TurnOffExperimental()
	})

	AfterEach(func() {
		helpers.TurnOnExperimental()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-cancel-zdt-push", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-cancel-zdt-push - Cancel the most recent deployment for an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-cancel-zdt-push APP_NAME"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		When("the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithAPIVersions(helpers.DefaultV2Version, "3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-cancel-zdt-push", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`This command requires CF API version 3\.27\.0 or higher\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-cancel-zdt-push", appName)
		})
	})

	Context("when the environment is set up correctly", func() {

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("v3-cancel-zdt-push")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		It("displays the experimental warning", func() {
			session := helpers.CF("v3-cancel-zdt-push", appName)
			Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
			Eventually(session).Should(Exit())
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-cancel-zdt-push", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No org targeted, use 'cf target -o ORG' to target an org\.`))
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
				session := helpers.CF("v3-cancel-zdt-push", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No space targeted, use 'cf target -s SPACE' to target a space\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-cancel-zdt-push", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No org targeted, use 'cf target -o ORG' to target an org\.`))
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
				session := helpers.CF("v3-cancel-zdt-push", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No space targeted, use 'cf target -s SPACE' to target a space\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-cancel-zdt-push", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No org targeted, use 'cf target -o ORG' to target an org\.`))
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
				session := helpers.CF("v3-cancel-zdt-push", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No space targeted, use 'cf target -s SPACE' to target a space\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app exists", func() {
			var session *Session

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)
					Eventually(session).Should(Exit(0))
				})
			})

			Context("and it is currently deploying", func() {
				BeforeEach(func() {
					// zdt-push a crashing app so that underlying deployment never goes from DEPLOYING -> DEPLOYED
					helpers.WithCrashingApp(func(appDir string) {
						session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-zdt-push", appName)
						Eventually(session).Should(Exit(0))
					})
				})

				It("cancels the deployment", func() {
					session := helpers.CF("v3-cancel-zdt-push", appName)
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say("Deployment cancelled, rolling back"))
				})
			})

			Context("when the app has no deployment to cancel", func() {
				It("errors", func() {
					session := helpers.CF("v3-cancel-zdt-push", appName)
					Eventually(session).Should(Exit(1))
					Eventually(session.Err).Should(Say("failed to find a deployment for that app"))
				})
			})
		})
	})
})
