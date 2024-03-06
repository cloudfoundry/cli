package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.NewAppName()
	})

	Describe("help", func() {
		It("shows usage", func() {
			session := helpers.CF("help", "delete")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("\\s+delete - Delete an app"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf delete APP_NAME \\[-r\\] \\[-f\\]"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("\\s+-f\\s+Force deletion without confirmation"))
			Eventually(session).Should(Say("\\s+-r\\s+Also delete any mapped routes"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("apps, scale, stop"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.UnrefactoredCheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "delete", "app-name")
		})
	})

	When("the environment is setup correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("delete")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app does not exist", func() {
			It("displays that the app does not exist", func() {
				session := helpers.CF("delete", appName, "-f")

				Eventually(session).Should(Say("Deleting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("App %s does not exist.", appName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete", appName, "-f", "-r")).Should(Exit(0))
			})

			When("the -f flag not is provided", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				When("the user enters the default input (hits return)", func() {
					BeforeEach(func() {
						_, err := buffer.Write([]byte("\n"))
						Expect(err).NotTo(HaveOccurred())
					})

					It("does not delete the app", func() {
						session := helpers.CFWithStdin(buffer, "delete", appName)
						Eventually(session).Should(Say("Really delete the app %s\\?", appName))
						Eventually(session).Should(Say("Delete cancelled"))
						Eventually(session).Should(Exit(0))
						Eventually(helpers.CF("app", appName)).Should(Exit(0))
					})
				})

			})

			When("the -f flag is provided", func() {
				It("deletes the app without prompting", func() {
					session := helpers.CF("delete", appName, "-f")
					Eventually(session).Should(Say("Deleting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("app", appName)).Should(Exit(1))
				})
			})
		})
	})
})
