package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with only an app name when authenticated with client-credentials", func() {
	var (
		appName  string
		clientID string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		clientID = helpers.LoginCFWithClientCredentials()
		helpers.TargetOrgAndSpace(organization, space)
	})

	AfterEach(func() {
		helpers.LogoutCF()
	})

	Describe("app existence", func() {
		When("the app does not exist", func() {
			It("creates the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, clientID))
					Eventually(session).Should(Say(`Getting app info\.\.\.`))
					Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
					Eventually(session).Should(Say(`\+\s+name:\s+%s`, appName))
					Eventually(session).Should(Say(`\s+routes:`))
					Eventually(session).Should(Say(`(?i)\+\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Say(`Mapping routes\.\.\.`))
					Eventually(session).Should(Say(`Comparing local files to remote cache\.\.\.`))
					Eventually(session).Should(Say(`Packaging files to upload\.\.\.`))
					Eventually(session).Should(Say(`Uploading files\.\.\.`))
					Eventually(session).Should(Say("100.00%"))
					Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
					helpers.ConfirmStagingLogs(session)
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
