package integration

import (
	"fmt"

	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("run-task command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		Skip("until bosh-lites are running CAPI V2.64")
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
		session := CF("run-task", "-h")
		Eventually(session).Should(Exit(0))
		Expect(session.Out).To(Say(`NAME:
   run-task - Run a one-off task on an app

USAGE:
   cf run-task APP_NAME COMMAND

EXAMPLES:
   cf run-task my-app "bundle exec rake db:migrate"

ALIAS:
   rt

SEE ALSO:
   tasks, terminate-task`))
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

		Context("when there no org set", func() {
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

		Context("when there no space set", func() {
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
		Context("when the application exists", func() {
			BeforeEach(func() {
				WithSimpleApp(func(appDir string) {
					Eventually(CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack"), CFLongTimeout).Should(Exit(0))
				})
			})

			It("creates a new task", func() {
				session := CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(0))
				userName, _ := getCredentials()
				Expect(session.Out).To(Say(fmt.Sprintf("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
				Expect(session.Out).To(Say(`OK

Task 1 has been submitted successfully for execution.`,
				))
			})
		})

		Context("when the application is not staged", func() {
			BeforeEach(func() {
				WithSimpleApp(func(appDir string) {
					Eventually(CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack"), CFLongTimeout).Should(Exit(0))
				})
			})

			It("fails and outputs task must have a droplet message", func() {
				session := CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say(`Unexpected Response
Response Code: 422
Code: 10008, Title: CF-UnprocessableEntity, Detail: The request is semantically invalid: Task must have a droplet. Specify droplet or assign current droplet to app`))
			})
		})

		Context("when the application is staged but stopped", func() {
			BeforeEach(func() {
				WithSimpleApp(func(appDir string) {
					Eventually(CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack"), CFLongTimeout).Should(Exit(0))
				})
				session := CF("stop", appName)
				Eventually(session).Should(Exit(0))
			})

			It("creates a new task", func() {
				session := CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(0))
				userName, _ := getCredentials()
				Expect(session.Out).To(Say(fmt.Sprintf("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
				Expect(session.Out).To(Say(`OK

Task 1 has been submitted successfully for execution.`,
				))
			})
		})

		Context("when the application does not exist", func() {
			It("fails and outputs an app not found message", func() {
				session := CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say(fmt.Sprintf("App %s not found", appName)))
			})
		})
	})
})
