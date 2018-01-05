package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = XDescribe("delete command", func() {
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

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "delete", "app-name")
		})
	})

	Context("when the environment is setup correctly", func() {
		var userName string

		BeforeEach(func() {
			setupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("delete")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app does not exist", func() {
			It("prompts the user and displays app does not exist", func() {
				buffer := NewBuffer()
				buffer.Write([]byte("y\n"))
				session := helpers.CFWithStdin(buffer, "delete")

				Eventually(session.Out).Should(Say("Really delete the app %s\\? \\[yN\\]", appName))
				Eventually(session.Out).Should(Say("Deleting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("App %s does not exist.", appName))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v2-push", appName)).Should(Exit(0))
				})
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete", appName, "-f", "-r")).Should(Exit(0))
			})

			Context("when the -f flag not is provided", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				Context("when the user enters 'y'", func() {
					BeforeEach(func() {
						buffer.Write([]byte("y\n"))
					})

					It("deletes the app", func() {
						session := helpers.CFWithStdin(buffer, "delete", appName)
						Eventually(session.Out).Should(Say("Really delete the app %s\\? \\[yN\\]", appName))
						Eventually(session.Out).Should(Say("Deleting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
						Eventually(helpers.CF("app", appName)).Should(Exit(1))
					})
				})

				Context("when the user enters 'n'", func() {
					BeforeEach(func() {
						buffer.Write([]byte("n\n"))
					})

					It("does not delete the app", func() {
						session := helpers.CFWithStdin(buffer, "delete", appName)
						Eventually(session.Out).Should(Say("Really delete the app %s\\? \\[yN\\]", appName))
						Eventually(session.Out).Should(Say("Delete cancelled"))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
						Eventually(helpers.CF("app", appName)).Should(Exit(0))
					})
				})

				Context("when the user enters the default input (hits return)", func() {
					BeforeEach(func() {
						buffer.Write([]byte("\n"))
					})

					It("does not delete the app", func() {
						session := helpers.CFWithStdin(buffer, "delete", appName)
						Eventually(session.Out).Should(Say("Really delete the app %s\\? \\[yN\\]", appName))
						Eventually(session.Out).Should(Say("Delete cancelled"))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
						Eventually(helpers.CF("app", appName)).Should(Exit(0))
					})
				})

				Context("when the user enters an invalid answer", func() {
					BeforeEach(func() {
						// The second '\n' is intentional. Otherwise the buffer will be
						// closed while the interaction is still waiting for input; it gets
						// an EOF and causes an error.
						buffer.Write([]byte("wat\n\n"))
					})

					It("asks again", func() {
						session := helpers.CFWithStdin(buffer, "delete", appName)
						Eventually(session.Out).Should(Say("Really delete the app %s\\? \\[yN\\]", appName))
						Eventually(session.Out).Should(Say("invalid input \\(not y, n, yes, or no\\)"))
						Eventually(session.Out).Should(Say("Really delete the app %s\\? \\[yN\\]", appName))
						Eventually(session).Should(Exit(0))
						Eventually(helpers.CF("app", appName)).Should(Exit(0))
					})
				})
			})

			Context("when the -f flag is provided", func() {
				It("deletes the app without prompting", func() {
					session := helpers.CF("delete", appName, "-f")
					Eventually(session.Out).Should(Say("Deleting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Delete cancelled"))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("app", appName)).Should(Exit(1))
				})
			})
		})
	})
})
