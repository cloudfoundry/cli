package push

import (
	"fmt"
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
		appName  string
		username string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		username, _ = helpers.GetCredentials()
	})

	Context("when the app is new", func() {
		Context("when the manifest is in the current directory", func() {
			Context("with no global properties", func() {
				It("uses the manifest for app settings", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name":       appName,
									"path":       dir,
									"command":    fmt.Sprintf("echo 'hi' && %s", helpers.StaticfileBuildpackStartCommand),
									"buildpack":  "staticfile_buildpack",
									"disk_quota": "300M",
									"env": map[string]interface{}{
										"key1": "val1",
										"key2": 2,
										"key3": true,
										"key4": 123412341234,
										"key5": 123456789.12345678,
									},
									"instances":                  2,
									"memory":                     "70M",
									"stack":                      "cflinuxfs2",
									"health-check-type":          "http",
									"health-check-http-endpoint": "/",
									"timeout":                    180,
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
						Eventually(session).Should(Say("Pushing from manifest to org %s / space %s as %s\\.\\.\\.", organization, space, username))
						Eventually(session).Should(Say("Getting app info\\.\\.\\."))
						Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
						Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
						Eventually(session).Should(Say("\\s+path:\\s+%s", regexp.QuoteMeta(dir)))
						Eventually(session).Should(Say("\\s+buildpack:\\s+staticfile_buildpack"))
						Eventually(session).Should(Say("\\s+command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
						Eventually(session).Should(Say("\\s+disk quota:\\s+300M"))
						Eventually(session).Should(Say("\\s+health check http endpoint:\\s+/"))
						Eventually(session).Should(Say("\\s+health check timeout:\\s+180"))
						Eventually(session).Should(Say("\\s+health check type:\\s+http"))
						Eventually(session).Should(Say("\\s+instances:\\s+2"))
						Eventually(session).Should(Say("\\s+memory:\\s+70M"))
						Eventually(session).Should(Say("\\s+stack:\\s+cflinuxfs2"))
						Eventually(session).Should(Say("\\s+env:"))
						Eventually(session).Should(Say("\\+\\s+key1"))
						Eventually(session).Should(Say("\\+\\s+key2"))
						Eventually(session).Should(Say("\\+\\s+key3"))
						Eventually(session).Should(Say("\\+\\s+key4"))
						Eventually(session).Should(Say("\\+\\s+key5"))
						Eventually(session).Should(Say("\\s+routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
						Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
						Eventually(session).Should(Say("Uploading files\\.\\.\\."))
						Eventually(session).Should(Say("100.00%"))
						Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
						helpers.ConfirmStagingLogs(session)
						Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
						Eventually(session).Should(Say("requested state:\\s+started"))
						Eventually(session).Should(Say("start command:\\s+echo 'hi' && %s", regexp.QuoteMeta(helpers.StaticfileBuildpackStartCommand)))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("instances:\\s+\\d/2"))
					Eventually(session).Should(Say("usage:\\s+70M x 2"))
					Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
					Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
					Eventually(session).Should(Say("#0.* of 70M"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", appName)
					Eventually(session).Should(Say("key1:\\s+val1"))
					Eventually(session).Should(Say("key2:\\s+2"))
					Eventually(session).Should(Say("key3:\\s+true"))
					Eventually(session).Should(Say("key4:\\s+123412341234"))
					Eventually(session).Should(Say("key5:\\s+123456789.12345678"))
					Eventually(session).Should(Exit(0))
				})

				Context("when health-check-type is http and no endpoint is provided", func() {
					It("defaults health-check-http-endpoint to '/'", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name":              appName,
										"path":              dir,
										"health-check-type": "http",
									},
								},
							})

							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
							Eventually(session).Should(Say("Getting app info\\.\\.\\."))
							Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
							Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
							Eventually(session).Should(Say("\\s+health check http endpoint:\\s+/"))
							Eventually(session).Should(Say("\\s+health check type:\\s+http"))
							Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
							Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
							Eventually(session).Should(Say("requested state:\\s+started"))
							Eventually(session).Should(Exit(0))
						})

						session := helpers.CF("app", appName)
						Eventually(session).Should(Exit(0))
					})
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

	Context("when the app already exists", func() {
		Context("when the app has manifest properties", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": appName,
								"path": dir,
								"env": map[string]interface{}{
									"key1": "val10",
									"key2": 2,
									"key3": true,
								},
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("\\s+env:"))
					Eventually(session).Should(Say("\\+\\s+key1"))
					Eventually(session).Should(Say("\\+\\s+key2"))
					Eventually(session).Should(Say("\\+\\s+key3"))
					Eventually(session).Should(Exit(0))
				})
			})

			It("adds or overrides the original env values", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": appName,
								"path": dir,
								"env": map[string]interface{}{
									"key1": "val1",
									"key4": false,
								},
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session).Should(Say("\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("\\s+env:"))
					Eventually(session).Should(Say("\\-\\s+key1"))
					Eventually(session).Should(Say("\\+\\s+key1"))
					Eventually(session).Should(Say("\\s+key2"))
					Eventually(session).Should(Say("\\s+key3"))
					Eventually(session).Should(Say("\\+\\s+key4"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("env", appName)
				Eventually(session).Should(Say("key1:\\s+val1"))
				Eventually(session).Should(Say("key2:\\s+2"))
				Eventually(session).Should(Say("key3:\\s+true"))
				Eventually(session).Should(Say("key4:\\s+false"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
