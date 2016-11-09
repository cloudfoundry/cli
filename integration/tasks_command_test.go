package integration

import (
	"fmt"

	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("tasks command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = PrefixedRandomName("ORG")
		spaceName = PrefixedRandomName("SPACE")
		appName = PrefixedRandomName("APP")

		setupCF(orgName, spaceName)
	})

	AfterEach(func() {
		setAPI()
		loginCF()
		Eventually(CF("delete-org", "-f", orgName), CFLongTimeout).Should(Exit(0))
	})

	It("should display the command level help", func() {
		session := CF("tasks", "-h")
		Eventually(session).Should(Exit(0))
		Expect(session.Out).To(Say(`NAME:
   tasks - List tasks of an app

USAGE:
   cf tasks APP_NAME

SEE ALSO:
   apps, run-task, terminate-task`,
		))
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				unsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := CF("run-task", "app-name", "some command")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				logoutCF()
			})

			It("fails with not logged in message", func() {
				session := CF("run-task", "app-name", "some command")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				logoutCF()
				loginCF()
			})

			It("fails with no targeted org error message", func() {
				session := CF("run-task", "app-name", "some command")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No org targeted, use 'cf target -o ORG' to target an org."))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				// create a another space, because if the org has only one space it
				// will be automatically targetted
				createSpace(PrefixedRandomName("SPACE"))
				logoutCF()
				loginCF()
				targetOrg(orgName)
			})

			It("fails with no space targeted error message", func() {
				session := CF("run-task", "app-name", "some command")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No space targeted, use 'cf target -s SPACE' to target a space"))
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		Context("when the application does not exist", func() {
			It("fails and outputs an app not found message", func() {
				session := CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say(fmt.Sprintf("App %s not found", appName)))
			})
		})

		Context("when the application exists", func() {
			BeforeEach(func() {
				WithSimpleApp(func(appDir string) {
					Eventually(CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack"), CFLongTimeout).Should(Exit(0))
				})
			})

			Context("when the application does not have associated tasks", func() {
				It("displays an empty table", func() {
					session := CF("tasks", appName)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say("id   name   state   start time   command\n"))
				})
			})

			Context("when the application has associated tasks", func() {
				BeforeEach(func() {
					Eventually(CF("run-task", appName, "echo hello world"), CFLongTimeout).Should(Exit(0))
					Eventually(CF("run-task", appName, "echo foo bar"), CFLongTimeout).Should(Exit(0))
				})

				It("displays all the tasks in descending order", func() {
					session := CF("tasks", appName)
					Eventually(session).Should(Exit(0))
					userName, _ := getCredentials()
					Expect(session.Out).To(Say(fmt.Sprintf("Getting tasks for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
					Expect(session.Out).To(Say("OK\n"))
					Expect(session.Out).To(Say(`id\s+name\s+state\s+start time\s+command
2\s+[a-zA-Z-0-9 ,:]+echo foo bar
1\s+[a-zA-Z-0-9 ,:]+echo hello world`))
				})

				Context("when the logged in user does not have authorization to see task commands", func() {
					var user string

					BeforeEach(func() {
						user = PrefixedRandomName("USER")
						password := PrefixedRandomName("PASSWORD")
						Eventually(CF("create-user", user, password), CFLongTimeout).Should(Exit(0))
						Eventually(CF("set-space-role", user, orgName, spaceName, "SpaceAuditor"), CFLongTimeout).Should(Exit(0))
						Eventually(CF("auth", user, password), CFLongTimeout).Should(Exit(0))
						Eventually(CF("target", "-o", orgName, "-s", spaceName), CFLongTimeout).Should(Exit(0))
					})

					AfterEach(func() {
						loginCF()
						Eventually(CF("delete-user", user, "-f"), CFLongTimeout).Should(Exit(0))
					})

					It("does not display task commands", func() {
						session := CF("tasks", appName)
						Eventually(session).Should(Exit(0))
						Expect(session.Out).To(Say("2\\s+[a-zA-Z-0-9 ,:]+\\[hidden\\]"))
						Expect(session.Out).To(Say("1\\s+[a-zA-Z-0-9 ,:]+\\[hidden\\]"))
					})
				})
			})
		})
	})
})
