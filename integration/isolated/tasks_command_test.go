package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("tasks command", func() {
	Context("when --help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("tasks", "--help")
			Eventually(session).Should(Exit(0))
			Expect(session.Out).To(Say("NAME:"))
			Expect(session.Out).To(Say("   tasks - List tasks of an app"))
			Expect(session.Out).To(Say("USAGE:"))
			Expect(session.Out).To(Say("   cf tasks APP_NAME"))
			Expect(session.Out).To(Say("SEE ALSO:"))
			Expect(session.Out).To(Say("   apps, logs, run-task, terminate-task"))
		})
	})
	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("run-task", "app-name", "some command")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("run-task", "app-name", "some command")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("run-task", "app-name", "some command")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No org targeted, use 'cf target -o ORG' to target an org."))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no space targeted error message", func() {
				session := helpers.CF("run-task", "app-name", "some command")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No space targeted, use 'cf target -s SPACE' to target a space"))
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		var (
			orgName   string
			spaceName string
			appName   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.PrefixedRandomName("SPACE")
			appName = helpers.PrefixedRandomName("APP")

			setupCF(orgName, spaceName)
		})

		Context("when the application does not exist", func() {
			It("fails and outputs an app not found message", func() {
				session := helpers.CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say(fmt.Sprintf("App %s not found", appName)))
			})
		})

		Context("when the application exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			Context("when the application does not have associated tasks", func() {
				It("displays an empty table", func() {
					session := helpers.CF("tasks", appName)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say(`
id   name   state   start time   command
`,
					))
					Expect(session.Out).NotTo(Say("1"))
				})
			})

			Context("when the application has associated tasks", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("run-task", appName, "echo hello world")).Should(Exit(0))
					Eventually(helpers.CF("run-task", appName, "echo foo bar")).Should(Exit(0))
				})

				It("displays all the tasks in descending order", func() {
					session := helpers.CF("tasks", appName)
					Eventually(session).Should(Exit(0))
					userName, _ := helpers.GetCredentials()
					Expect(session.Out).To(Say(fmt.Sprintf("Getting tasks for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
					Expect(session.Out).To(Say("OK\n"))
					Expect(session.Out).To(Say(`id\s+name\s+state\s+start time\s+command
2\s+[a-zA-Z-0-9 ,:]+echo foo bar
1\s+[a-zA-Z-0-9 ,:]+echo hello world`))
				})

				Context("when the logged in user does not have authorization to see task commands", func() {
					var user string

					BeforeEach(func() {
						user = helpers.RandomUsername()
						password := helpers.RandomPassword()
						Eventually(helpers.CF("create-user", user, password)).Should(Exit(0))
						Eventually(helpers.CF("set-space-role", user, orgName, spaceName, "SpaceAuditor")).Should(Exit(0))
						Eventually(helpers.CF("auth", user, password)).Should(Exit(0))
						Eventually(helpers.CF("target", "-o", orgName, "-s", spaceName)).Should(Exit(0))
					})

					It("does not display task commands", func() {
						session := helpers.CF("tasks", appName)
						Eventually(session).Should(Exit(0))
						Expect(session.Out).To(Say("2\\s+[a-zA-Z-0-9 ,:]+\\[hidden\\]"))
						Expect(session.Out).To(Say("1\\s+[a-zA-Z-0-9 ,:]+\\[hidden\\]"))
					})
				})
			})
		})
	})
})
