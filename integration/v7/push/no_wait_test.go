package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with --no-wait", func() {
	var (
		appName  string
		userName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
		userName, _ = helpers.GetCredentials()
	})

	When("the app exists", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName,
				)).Should(Exit(0))
			})
		})

		It("pushes the app", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName, "--no-wait",
				)

				Eventually(session).Should(Say(`Updating app %s\.\.\.`, appName))
				Eventually(session).Should(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`Getting app info\.\.\.`))
				Eventually(session).Should(Say(`Packaging files to upload\.\.\.`))
				Eventually(session).Should(Say(`Uploading files\.\.\.`))
				Eventually(session).Should(Say(`100.00%`))
				Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
				Eventually(session).Should(Say(`Staging app and tracing logs\.\.\.`))
				Eventually(session).Should(Say(`Waiting for app %s to start\.\.\.`, appName))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`start command:\s+%s`, helpers.StaticfileBuildpackStartCommand))
				Eventually(session).Should(Say(`#0\s+running`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app does not exist", func() {
		It("pushes the app", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName, "--no-wait",
				)

				Eventually(session).Should(Say(`Creating app %s\.\.\.`, appName))
				Eventually(session).Should(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`Getting app info\.\.\.`))
				Eventually(session).Should(Say(`Packaging files to upload\.\.\.`))
				Eventually(session).Should(Say(`Uploading files\.\.\.`))
				Eventually(session).Should(Say(`100.00%`))
				Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
				Eventually(session).Should(Say(`Staging app and tracing logs\.\.\.`))
				Eventually(session).Should(Say(`Waiting for app %s to start\.\.\.`, appName))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`start command:\s+%s`, helpers.StaticfileBuildpackStartCommand))
				Eventually(session).Should(Say(`#0\s+running`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app crashes", func() {
		It("times out", func() {
			helpers.WithCrashingApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{
					WorkingDirectory: appDir,
					EnvVars:          map[string]string{"CF_STARTUP_TIMEOUT": "0.1"},
				}, PushCommandName, appName, "--no-wait")
				Eventually(session).Should(Say(`Creating app %s\.\.\.`, appName))
				Eventually(session).Should(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`Getting app info\.\.\.`))
				Eventually(session).Should(Say(`Packaging files to upload\.\.\.`))
				Eventually(session).Should(Say(`Uploading files\.\.\.`))
				Eventually(session).Should(Say(`100.00%`))
				Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
				Eventually(session).Should(Say(`Staging app and tracing logs\.\.\.`))
				Eventually(session).Should(Say(`Waiting for app %s to start\.\.\.`, appName))
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Start app timeout`))
				Eventually(session.Err).Should(Say(`TIP: Application must be listening on the right port\. Instead of hard coding the port, use the \$PORT environment variable\.`))
				Eventually(session.Err).Should(Say(`Use 'cf logs %s --recent' for more information`, appName))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
