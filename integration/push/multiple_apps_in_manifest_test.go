package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushes multiple apps with a single manifest file", func() {
	var (
		firstApp  string
		secondApp string
	)

	BeforeEach(func() {
		firstApp = helpers.NewAppName()
		secondApp = helpers.NewAppName()
	})

	Context("when the apps are new", func() {
		Context("with no global properties", func() {
			It("pushes multiple apps with a single push", func() {
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

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))

					// firstApp
					Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", firstApp))
					Eventually(session).Should(Say("\\s+routes:"))
					Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", firstApp, defaultSharedDomain()))

					// secondApp
					Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", secondApp))
					Eventually(session).Should(Say("\\s+routes:"))
					Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", secondApp, defaultSharedDomain()))

					Eventually(session).Should(Say("Creating app %s\\.\\.\\.", firstApp))
					Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Say("Uploading files\\.\\.\\."))
					Eventually(session).Should(Say("100.00%"))
					Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
					helpers.ConfirmStagingLogs(session)
					Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
					Eventually(session).Should(Say("requested state:\\s+started"))

					Eventually(session).Should(Say("Creating app %s\\.\\.\\.", secondApp))
					Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Say("Uploading files\\.\\.\\."))
					Eventually(session).Should(Say("100.00%"))
					Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
					helpers.ConfirmStagingLogs(session)
					Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", firstApp)
				Eventually(session).Should(Say("name:\\s+%s", firstApp))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", secondApp)
				Eventually(session).Should(Say("name:\\s+%s", secondApp))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
