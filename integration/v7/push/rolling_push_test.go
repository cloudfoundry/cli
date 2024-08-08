package push

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with --strategy rolling", func() {
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

		It("pushes the app and creates a new deployment", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName, "--strategy", "rolling",
				)

				Eventually(session).Should(Exit(0))
				Expect(session).To(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Expect(session).To(Say(`Packaging files to upload\.\.\.`))
				Expect(session).To(Say(`Uploading files\.\.\.`))
				Expect(session).To(Say(`100.00%`))
				Expect(session).To(Say(`Waiting for API to complete processing files\.\.\.`))
				Expect(session).To(Say(`Staging app and tracing logs\.\.\.`))
				Expect(session).To(Say(`Starting deployment for app %s\.\.\.`, appName))
				Expect(session).To(Say(`Waiting for app to deploy\.\.\.`))
				Expect(session).To(Say(`name:\s+%s`, appName))
				Expect(session).To(Say(`requested state:\s+started`))
				Expect(session).To(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
				Expect(session).To(Say(`type:\s+web`))
				Expect(session).To(Say(`start command:\s+%s`, helpers.StaticfileBuildpackStartCommand))
				Expect(session).To(Say(`#0\s+running`))
			})
		})

		It("pushes the app and creates a new deployment with max in flight set", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName, "--strategy", "rolling", "--max-in-flight", "3",
				)

				Eventually(session).Should(Exit(0))
				Expect(session).To(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Expect(session).To(Say(`Packaging files to upload\.\.\.`))
				Expect(session).To(Say(`Uploading files\.\.\.`))
				Expect(session).To(Say(`100.00%`))
				Expect(session).To(Say(`Waiting for API to complete processing files\.\.\.`))
				Expect(session).To(Say(`Staging app and tracing logs\.\.\.`))
				Expect(session).To(Say(`Starting deployment for app %s\.\.\.`, appName))
				Expect(session).To(Say(`Waiting for app to deploy\.\.\.`))
				Expect(session).To(Say(`name:\s+%s`, appName))
				Expect(session).To(Say(`requested state:\s+started`))
				Expect(session).To(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
				Expect(session).To(Say(`type:\s+web`))
				Expect(session).To(Say(`start command:\s+%s`, helpers.StaticfileBuildpackStartCommand))
				Expect(session).To(Say(`#0\s+running`))
			})
		})
	})

	When("canceling the deployment", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName,
				)).Should(Exit(0))
			})
		})

		It("displays the deployment cancellation message", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName, "--strategy", "rolling",
				)

				Eventually(session).Should(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`Packaging files to upload\.\.\.`))
				Eventually(session).Should(Say(`Uploading files\.\.\.`))
				Eventually(session).Should(Say(`100.00%`))
				Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
				Eventually(session).Should(Say(`Staging app and tracing logs\.\.\.`))
				Eventually(session).Should(Say(`Starting deployment for app %s\.\.\.`, appName))
				Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))

				Eventually(helpers.CF("cancel-deployment", appName)).Should(Exit(0))
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Deployment has been canceled`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the app crashes", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName,
				)).Should(Exit(0))
			})
		})

		It("times out", func() {
			helpers.WithCrashingApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{
					WorkingDirectory: appDir,
					EnvVars:          map[string]string{"CF_STARTUP_TIMEOUT": "0.1"},
				}, PushCommandName, appName, "--strategy", "rolling")
				Eventually(session).Should(Exit(1))
				Expect(session).To(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Expect(session).To(Say(`Packaging files to upload\.\.\.`))
				Expect(session).To(Say(`Uploading files\.\.\.`))
				Expect(session).To(Say(`100.00%`))
				Expect(session).To(Say(`Waiting for API to complete processing files\.\.\.`))
				Expect(session).To(Say(`Staging app and tracing logs\.\.\.`))
				Expect(session).To(Say(`Starting deployment for app %s\.\.\.`, appName))
				Expect(session).To(Say(`Waiting for app to deploy\.\.\.`))
				Expect(session).To(Say(`FAILED`))
				Expect(session.Err).To(Say(`Start app timeout`))
				Expect(session.Err).To(Say(`TIP: Application must be listening on the right port\. Instead of hard coding the port, use the \$PORT environment variable\.`))
				Expect(session.Err).To(Say(`Use 'cf logs %s --recent' for more information`, appName))
				appGUID := helpers.AppGUID(appName)
				Eventually(func() *Buffer {
					session_deployment := helpers.CF("curl", fmt.Sprintf("/v3/deployments?app_guids=%s", appGUID))
					Eventually(session_deployment).Should(Exit(0))
					return session_deployment.Out
				}).Should(Say(`"reason":\s*"CANCELED"`))
			})
		})
	})
})
