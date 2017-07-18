package push

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with a simple manifest and flags", func() {
	var (
		appName        string
		pathToManifest string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()

		tmpFile, err := ioutil.TempFile("", "combination-manifest")
		Expect(err).ToNot(HaveOccurred())
		pathToManifest = tmpFile.Name()
		Expect(tmpFile.Close()).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(pathToManifest)).ToNot(HaveOccurred())
	})

	Context("when the app is new", func() {
		Context("when the manifest is passed via '-f'", func() {
			BeforeEach(func() {
				helpers.WriteManifest(pathToManifest, map[string]interface{}{
					"applications": []map[string]string{
						{
							"name": appName,
						},
					},
				})
			})

			It("uses the manifest for app settings", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", pathToManifest)
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

		Context("manifest contains a path and a '-p' is provided", func() {
			It("overrides the manifest path with the '-p' path", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]string{
							{
								"name": appName,
								"path": "does-not-exist",
							},
						},
					})

					session := helpers.CF(PushCommandName, "-p", dir)
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("\\s+path:\\s+%s", regexp.QuoteMeta(dir)))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the --no-manifest flag is passed", func() {
			It("does not use the provided manifest", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]string{
							{
								"name": "crazy-jerry",
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-manifest", appName)
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
	})
})
