package push

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with a simple manifest and flags", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when the app is new", func() {
		Context("when pushing a single app from the manifest", func() {
			Context("when the '-f' flag is provided", func() {
				var (
					pathToManifest string // Can be a filepath or a directory with a manifest.
				)

				Context("when the manifest file is passed", func() {
					BeforeEach(func() {
						tmpFile, err := ioutil.TempFile("", "combination-manifest")
						Expect(err).ToNot(HaveOccurred())
						pathToManifest = tmpFile.Name()
						Expect(tmpFile.Close()).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						Expect(os.Remove(pathToManifest)).ToNot(HaveOccurred())
					})

					Context("when pushing the app from the current directory", func() {
						BeforeEach(func() {
							helpers.WriteManifest(pathToManifest, map[string]interface{}{
								"applications": []map[string]string{
									{
										"name": appName,
									},
								},
							})
						})

						It("pushes the app from the current directory and the manifest for app settings", func() {
							helpers.WithHelloWorldApp(func(dir string) {
								session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", pathToManifest)
								Eventually(session).Should(Say("Getting app info\\.\\.\\."))
								Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
								Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
								Eventually(session).Should(Say("\\s+routes:"))
								Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
								Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
								Eventually(session).Should(Say("Uploading files\\.\\.\\."))
								Eventually(session).Should(Say("100.00%"))
								Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
								helpers.ConfirmStagingLogs(session)
								Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
								Eventually(session).Should(Say("requested state:\\s+started"))
								Eventually(session).Should(Exit(0))
							})

							session := helpers.CF("app", appName)
							Eventually(session).Should(Say("name:\\s+%s", appName))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("when the path to the application is provided in the manifest", func() {
						It("pushes the app from the path specified in the manifest and uses the manifest for app settings", func() {
							helpers.WithHelloWorldApp(func(dir string) {
								helpers.WriteManifest(pathToManifest, map[string]interface{}{
									"applications": []map[string]string{
										{
											"name": appName,
											"path": filepath.Base(dir),
										},
									},
								})

								session := helpers.CF(PushCommandName, "-f", pathToManifest)
								Eventually(session).Should(Say("Getting app info\\.\\.\\."))
								Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
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
				})

				Context("when a directory is passed", func() {
					var (
						ymlFile  string
						yamlFile string
					)

					BeforeEach(func() {
						var err error
						pathToManifest, err = ioutil.TempDir("", "manifest-integration-")
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						Expect(os.RemoveAll(pathToManifest)).ToNot(HaveOccurred())
					})

					Context("when the directory contains a 'manifest.yml' file", func() {
						BeforeEach(func() {
							ymlFile = filepath.Join(pathToManifest, "manifest.yml")
							helpers.WriteManifest(ymlFile, map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name":      appName,
										"instances": 2,
									},
								},
							})
						})

						It("pushes the app from the given directory and the found 'manifest.yml' for app settings", func() {
							helpers.WithHelloWorldApp(func(dir string) {
								session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", pathToManifest, "--no-start")
								Eventually(session).Should(Say("Using manifest file %s", regexp.QuoteMeta(ymlFile)))
								Eventually(session).Should(Say("Getting app info\\.\\.\\."))
								Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
								Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
								Eventually(session).Should(Say("\\+\\s+instances:\\s+%d", 2))
								Eventually(session).Should(Say("\\s+routes:"))
								Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
								Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
								Eventually(session).Should(Say("Uploading files\\.\\.\\."))
								Eventually(session).Should(Say("100.00%"))
								Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
								Eventually(session).Should(Exit(0))
							})

							session := helpers.CF("app", appName)
							Eventually(session).Should(Say("name:\\s+%s", appName))
							Eventually(session).Should(Exit(0))
						})
					})
					Context("when the directory contains a 'manifest.yaml' file", func() {
						BeforeEach(func() {
							yamlFile = filepath.Join(pathToManifest, "manifest.yaml")
							helpers.WriteManifest(yamlFile, map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name":      appName,
										"instances": 2,
									},
								},
							})
						})

						It("pushes the app from the given directory and the found 'manifest.yaml' for app settings", func() {
							helpers.WithHelloWorldApp(func(dir string) {
								session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", pathToManifest, "--no-start")
								Eventually(session).Should(Say("Using manifest file %s", regexp.QuoteMeta(yamlFile)))
								Eventually(session).Should(Say("Getting app info\\.\\.\\."))
								Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
								Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
								Eventually(session).Should(Say("\\+\\s+instances:\\s+%d", 2))
								Eventually(session).Should(Say("\\s+routes:"))
								Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
								Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
								Eventually(session).Should(Say("Uploading files\\.\\.\\."))
								Eventually(session).Should(Say("100.00%"))
								Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
								Eventually(session).Should(Exit(0))
							})

							session := helpers.CF("app", appName)
							Eventually(session).Should(Say("name:\\s+%s", appName))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("when the directory contains both a 'manifest.yml' file and a 'manifest.yaml' file", func() {
						BeforeEach(func() {
							ymlFile = filepath.Join(pathToManifest, "manifest.yml")
							helpers.WriteManifest(ymlFile, map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name":      appName,
										"instances": 2,
									},
								},
							})

							yamlFile = filepath.Join(pathToManifest, "manifest.yaml")
							helpers.WriteManifest(yamlFile, map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name":      appName,
										"instances": 4,
									},
								},
							})
						})

						It("pushes the app from the given directory and the found 'manifest.yml' for app settings", func() {
							helpers.WithHelloWorldApp(func(dir string) {
								session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", pathToManifest, "--no-start")
								Eventually(session).Should(Say("Using manifest file %s", regexp.QuoteMeta(ymlFile)))
								Eventually(session).Should(Say("Getting app info\\.\\.\\."))
								Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
								Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
								Eventually(session).Should(Say("\\+\\s+instances:\\s+%d", 2))
								Eventually(session).Should(Say("\\s+routes:"))
								Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
								Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
								Eventually(session).Should(Say("Uploading files\\.\\.\\."))
								Eventually(session).Should(Say("100.00%"))
								Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
								Eventually(session).Should(Exit(0))
							})

							session := helpers.CF("app", appName)
							Eventually(session).Should(Say("name:\\s+%s", appName))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("when the directory contains no manifest file", func() {
						It("returns a no manifest file error", func() {
							helpers.WithHelloWorldApp(func(dir string) {
								session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", pathToManifest, "--no-start")
								Eventually(session.Err).Should(Say("Could not find 'manifest\\.yml' file in %s", regexp.QuoteMeta(pathToManifest)))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session).Should(Exit(1))
							})
						})

					})
				})
			})

			Context("manifest contains a path and a '-p' is provided", func() {
				var tempDir string

				BeforeEach(func() {
					var err error
					tempDir, err = ioutil.TempDir("", "combination-manifest-with-p")
					Expect(err).ToNot(HaveOccurred())

					helpers.WriteManifest(filepath.Join(tempDir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]string{
							{
								"name": appName,
								"path": "does-not-exist",
							},
						},
					})
				})

				AfterEach(func() {
					Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
				})

				It("overrides the manifest path with the '-p' path", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, PushCommandName, "-p", dir)
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

			Context("manifest contains a name and a name is provided", func() {
				It("overrides the manifest name", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]string{
								{
									"name": "earle",
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
						Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
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
						Eventually(session).Should(Say("\\s+routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
						Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
						Eventually(session).Should(Say("Uploading files\\.\\.\\."))
						Eventually(session).Should(Say("100.00%"))
						Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
						helpers.ConfirmStagingLogs(session)
						Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
						Eventually(session).Should(Say("requested state:\\s+started"))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session.Out).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the manifest contains 'routes'", func() {
				var manifestContents map[string]interface{}

				BeforeEach(func() {
					manifestContents = map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": appName,
								"routes": []map[string]string{
									{"route": "some-route-1"},
									{"route": "some-route-2"},
								},
							},
						},
					}
				})

				Context("when the -d flag is provided", func() {
					It("returns an error message and exits 1", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifestContents)
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-d", "some-domain.com")

							Eventually(session).ShouldNot(Say("Getting app info"))
							Eventually(session.Err).Should(Say("The following arguments cannot be used with an app manifest that declares routes using the 'route' attribute: -d, --hostname, -n, --no-hostname, --route-path"))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("when the --hostname flag is provided", func() {
					It("returns an error message and exits 1", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifestContents)
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--hostname", "some-host")

							Eventually(session).ShouldNot(Say("Getting app info"))
							Eventually(session.Err).Should(Say("The following arguments cannot be used with an app manifest that declares routes using the 'route' attribute: -d, --hostname, -n, --no-hostname, --route-path"))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("when the -n flag is provided", func() {
					It("returns an error message and exits 1", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifestContents)
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-n", "some-host")

							Eventually(session).ShouldNot(Say("Getting app info"))
							Eventually(session.Err).Should(Say("The following arguments cannot be used with an app manifest that declares routes using the 'route' attribute: -d, --hostname, -n, --no-hostname, --route-path"))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("when the --no-hostname flag is provided", func() {
					It("returns an error message and exits 1", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifestContents)
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-hostname")

							Eventually(session).ShouldNot(Say("Getting app info"))
							Eventually(session.Err).Should(Say("The following arguments cannot be used with an app manifest that declares routes using the 'route' attribute: -d, --hostname, -n, --no-hostname, --route-path"))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("when the --route-path flag is provided", func() {
					It("returns an error message and exits 1", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifestContents)
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--route-path", "some-path")

							Eventually(session).ShouldNot(Say("Getting app info"))
							Eventually(session.Err).Should(Say("The following arguments cannot be used with an app manifest that declares routes using the 'route' attribute: -d, --hostname, -n, --no-hostname, --route-path"))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})
		})

		Context("when pushing multiple apps from the manifest", func() {
			Context("manifest contains multiple apps and '--no-start' is provided", func() {
				var appName1, appName2 string

				BeforeEach(func() {
					appName1 = helpers.NewAppName()
					appName2 = helpers.NewAppName()
				})

				It("does not start the apps", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]string{
								{"name": appName1},
								{"name": appName2},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
						Eventually(session).Should(Say("Getting app info\\.\\.\\."))
						Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
						Eventually(session).Should(Say("\\s+name:\\s+%s", appName1))
						Eventually(session).Should(Say("requested state:\\s+stopped"))
						Eventually(session).Should(Say("\\s+name:\\s+%s", appName2))
						Eventually(session).Should(Say("requested state:\\s+stopped"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("manifest contains multiple apps and a '-p' is provided", func() {
				var tempDir string

				BeforeEach(func() {
					var err error
					tempDir, err = ioutil.TempDir("", "combination-manifest-with-p")
					Expect(err).ToNot(HaveOccurred())

					helpers.WriteManifest(filepath.Join(tempDir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]string{
							{
								"name": "name-1",
							},
							{
								"name": "name-2",
							},
						},
					})
				})

				AfterEach(func() {
					Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
				})

				It("returns an error", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, PushCommandName, "-p", dir)
						Eventually(session.Err).Should(Say(regexp.QuoteMeta("Incorrect Usage: Command line flags (except -f and --no-start) cannot be applied when pushing multiple apps from a manifest file.")))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			DescribeTable("errors when any flag (except for -f and --no-start) is specified",
				func(flags ...string) {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]string{
								{"name": "some-app"},
								{"name": "some-other-app"},
							},
						})

						args := append([]string{PushCommandName}, flags...)
						session := helpers.CustomCF(helpers.CFEnv{
							WorkingDirectory: dir,
							EnvVars:          map[string]string{"CF_DOCKER_PASSWORD": "some-password"},
						}, args...)
						Eventually(session.Err).Should(Say(regexp.QuoteMeta("Incorrect Usage: Command line flags (except -f and --no-start) cannot be applied when pushing multiple apps from a manifest file.")))
						Eventually(session).Should(Exit(1))
					})
				},
				Entry("buildpack", "-b", "somethin"),
				Entry("domain", "-d", "something"),
				Entry("hostname", "-n", "something"),
				Entry("quota", "-k", "100M"),
				Entry("docker image", "-o", "something"),
				Entry("docker image and username", "-o", "something", "--docker-username", "something"),
				Entry("health check timeout", "-t", "10"),
				Entry("health check type", "-u", "http"),
				Entry("instances", "-i", "10"),
				Entry("memory", "-m", "100M"),
				Entry("no hostname", "--no-hostname"),
				Entry("no route", "--no-route"),
				Entry("random route", "--random-route"),
				Entry("route path", "--route-path", "something"),
				Entry("stack", "-s", "something"),
			)
		})
	})
})
