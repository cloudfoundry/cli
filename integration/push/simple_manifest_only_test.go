package push

import (
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with a simple manifest and no flags", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when the app is new", func() {
		Context("when the manifest is in the current directory", func() {
			Context("with no global properties", func() {
				It("uses the manifest for app settings", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]string{
								{
									"name": appName,
									"path": dir,
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
						Eventually(session).Should(Say("Getting app info\\.\\.\\."))
						Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
						Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
						Eventually(session).Should(Say("\\s+path:\\s+%s", regexp.QuoteMeta(dir)))
						Eventually(session).Should(Say("\\s+routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
						Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
						Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
						Eventually(session).Should(Say("Uploading files\\.\\.\\."))
						Eventually(session).Should(Say("100.00%"))
						Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
						Eventually(session).Should(Say("Downloaded staticfile_buildpack"))
						Eventually(session).Should(Say("Staging complete"))
						Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
						Eventually(session).Should(Say("requested state:\\s+started"))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app has no name", func() {
				It("returns an error", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]string{
								{
									"name": "",
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
						Eventually(session.Err).Should(Say("Incorrect usage: The push command requires an app name. The app name can be supplied as an argument or with a manifest.yml file."))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the app path does not exist", func() {
				It("returns an error", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]string{
								{
									"name": "some-name",
									"path": "does-not-exist",
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
						Eventually(session.Err).Should(Say("File not found locally, make sure the file exists at given path .*does-not-exist"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		Context("there is no name or no manifest", func() {
			It("returns an error", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
					Eventually(session.Err).Should(Say("Incorrect usage: The push command requires an app name. The app name can be supplied as an argument or with a manifest.yml file."))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
