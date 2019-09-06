package push

import (
	"path/filepath"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Output", func() {
	var (
		appName  string
		appName2 string

		username string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		appName2 = helpers.NewAppName()
		username, _ = helpers.GetCredentials()
	})

	When("there is a manifest", func() {
		It("prints the manifests path and other output", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{"name": appName},
						{"name": appName2},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
				Eventually(session).Should(Say(`Pushing apps %s, %s to org %s / space %s as %s\.\.\.`, appName, appName2, organization, space, username))
				Eventually(session).Should(helpers.SayPath(`Applying manifest file %s`, filepath.Join(dir, "manifest.yml")))
				Eventually(session).Should(Say(`Uploading files\.\.\.`))
				Eventually(session).Should(Say("100.00%"))
				Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
				helpers.ConfirmStagingLogs(session)
				Eventually(session).Should(Say(`Waiting for app %s to start\.\.\.`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`start command:\s+%s`, regexp.QuoteMeta(helpers.ModernStaticfileBuildpackStartCommand)))
				Eventually(session).Should(Say(`Uploading files\.\.\.`))
				Eventually(session).Should(Say("100.00%"))
				Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
				helpers.ConfirmStagingLogs(session)
				Eventually(session).Should(Say(`Waiting for app %s to start\.\.\.`, appName2))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`start command:\s+%s`, regexp.QuoteMeta(helpers.ModernStaticfileBuildpackStartCommand)))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("there is no a manifest", func() {
		It("does not display any manifest information", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
				Consistently(session).ShouldNot(Say(`Applying manifest`))
				Consistently(session, 10*time.Second).ShouldNot(Say(`Manifest applied`))
				Eventually(session).Should(Say(`Pushing app %s to org %s / space %s as %s\.\.\.`, appName, organization, space, username))
				Eventually(session).Should(Say(`Uploading files\.\.\.`))
				Eventually(session).Should(Say("100.00%"))
				Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
				helpers.ConfirmStagingLogs(session)
				Eventually(session).Should(Say(`Waiting for app %s to start\.\.\.`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`start command:\s+%s`, regexp.QuoteMeta(helpers.ModernStaticfileBuildpackStartCommand)))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
