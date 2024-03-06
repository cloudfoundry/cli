package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushes specified app from single manifest file", func() {
	var (
		firstApp  string
		secondApp string
	)

	BeforeEach(func() {
		firstApp = helpers.NewAppName()
		secondApp = helpers.NewAppName()
	})

	When("the specified app is not found in the manifest file", func() {
		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]string{
						{
							"name": firstApp,
						},
						{
							"name": secondApp,
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "some-app-not-from-manifest")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Could not find app named 'some-app-not-from-manifest' in manifest"))

				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the specified app exists in the manifest file", func() {
		It("pushes just the app on the command line", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]string{
						{
							"name": firstApp,
						},
						{
							"name": secondApp,
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, firstApp)
				Eventually(session).Should(Say(`Getting app info\.\.\.`))

				Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
				Eventually(session).Should(Say(`\+\s+name:\s+%s`, firstApp))
				Eventually(session).Should(helpers.SayPath(`path:\s+%s`, dir))
				Eventually(session).Should(Say(`\s+routes:`))
				Eventually(session).Should(Say(`(?i)\+\s+%s.%s`, firstApp, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say(`Uploading files\.\.\.`))
				Eventually(session).Should(Say("100.00%"))
				Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
				helpers.ConfirmStagingLogs(session)
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`requested state:\s+started`))

				Consistently(session).ShouldNot(Say(`\+\s+name:\s+%s`, secondApp))
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("app", firstApp)
			Eventually(session).Should(Say(`name:\s+%s`, firstApp))
			Eventually(session).Should(Exit(0))

			session = helpers.CF("app", secondApp)
			Eventually(session).ShouldNot(Say(`name:\s+%s`, secondApp))
			Eventually(session).Should(Exit(1))
		})
	})
})
