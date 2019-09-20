package push

import (
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"path/filepath"
)

var _ = Describe("push with health check type", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("updating the application", func() {
		BeforeEach(func() {
			helpers.WithMultiEndpointApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "http", "--endpoint", "/other_endpoint.html")
				Eventually(session).Should(Exit(0))
			})
		})

		When("setting the app to http health check type", func() {
			It("should update the health check http endpoint", func() {
				helpers.WithMultiEndpointApp(func(dir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "http", "--endpoint", "/third_endpoint.html")).Should(Exit(0))
				})

				session := helpers.CF("get-health-check", appName)
				Eventually(session).Should(Say("web\\s+http\\s+/third_endpoint.html"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("setting the app to port health check type", func() {
			It("should reset the health check http endpoint", func() {
				helpers.WithMultiEndpointApp(func(dir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "port")).Should(Exit(0))
				})

				session := helpers.CF("get-health-check", appName)
				Eventually(session).Should(Say("web\\s+port\\s+1"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("setting the app to process health check type", func() {
			It("should reset the health check http endpoint", func() {
				helpers.WithMultiEndpointApp(func(dir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "process")).Should(Exit(0))
				})

				session := helpers.CF("get-health-check", appName)
				Eventually(session).Should(Say("web\\s+process\\s+1"))
				Eventually(session).Should(Exit(0))
			})
		})

		FWhen("setting the app's health check http endpoint", func() {
			When("manifest has non-http (e.g. port) health check type", func() {

				// A/C # 1
				When("manifest has http health check type", func() {
					When("only the endpoint flag override is specified", func() {
						It("should update the endpoint", func() {
							helpers.WithMultiEndpointApp(func(dir string) {
								helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
									"applications": []map[string]string{
										{
											"name":                  "some-app",
											"health-check-type":     "http",
											"health-check-endpoint": "/",
										},
									},
								})

								session := helpers.CustomCF(
									helpers.CFEnv{WorkingDirectory: dir},
									PushCommandName, appName,
									"--endpoint", "/third_endpoint.html")
								Eventually(session).Should(Exit(0))
							})

							session := helpers.CF("get-health-check", appName)
							Eventually(session).Should(Say("web\\s+http\\s+/third_endpoint.html"))
							Eventually(session).Should(Exit(0)) //})
						})
					})

					When("both health check type and endpoint flag overrides are specified", func() {
						// A/C # 2
						When("only the endpoint flag override is specified", func() {
							It("should fail with an error indicating that endpoint cannot be set for health check type port", func() {
								helpers.WithMultiEndpointApp(func(dir string) {
									helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
										"applications": []map[string]string{
											{
												"name":                  "some-app",
												"health-check-type":     "http",
												"health-check-endpoint": "/",
											},
										},
									})

									session := helpers.CustomCF(
										helpers.CFEnv{WorkingDirectory: dir},
										PushCommandName, appName,
										"--endpoint", "/third_endpoint.html")
									// Because we no longer require that override --endpoint must be sent with override
									// --health-check-type=http
									// If manifest health-check-type is non-http
									// We will transform the manifest to have this combo:
									//   health-check-type: port # i.e. non-http
									//   health-check-endpoint: /health # i.e. any real endpoint on the app
									// This will then fail with a CAPI error when CAPI apply manifest endpoint complains
									//
									//Eventually(session).Should(Say("[error message about a conflict about how endpoints can't be set for port type health checks]"))
									Eventually(session).Should(Exit(1))
								})
							})
						})

						// A/C # 3
						It("should update the health check type and endpoint", func() {
							helpers.WithMultiEndpointApp(func(dir string) {
								helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
									"applications": []map[string]string{
										{
											"name":              "some-app",
											"health-check-type": "port",
										},
									},
								})

								session := helpers.CustomCF(
									helpers.CFEnv{WorkingDirectory: dir},
									PushCommandName, appName,
									"--health-check-type", "http",
									"--endpoint", "/third_endpoint.html")
								Eventually(session).Should(Exit(0))
							})

							session := helpers.CF("get-health-check", appName)
							Eventually(session).Should(Say("web\\s+http\\s+/third_endpoint.html"))
							Eventually(session).Should(Exit(0))
						})
					})

				})
			})
		})

		Context("creating the application", func() {
			When("setting a http health check type", func() {
				It("should set the health check type to http and use the default health check endpoint", func() {
					helpers.WithMultiEndpointApp(func(dir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "http", "--endpoint", "/other_endpoint.html")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session).Should(Say("web\\s+http\\s+/other_endpoint.html"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("setting a port health check type", func() {
				It("it should set the health check type to port", func() {
					helpers.WithMultiEndpointApp(func(dir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "port")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session).Should(Say("web\\s+port"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("setting a process health check type", func() {
				It("it should set the health check type to process", func() {
					helpers.WithMultiEndpointApp(func(dir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "process")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session).Should(Say("web\\s+process"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	When("there is a manifest", func() {
		var (
			pathToManifest string
		)
		BeforeEach(func() {
			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			pathToManifest = filepath.Join(tempDir, "manifest.yml")

			helpers.WriteManifest(pathToManifest, map[string]interface{}{
				"applications": []map[string]interface{}{{
					"name":                       "charlie",
					"health-check-type":          "http",
					"health-check-http-endpoint": "/",
				}},
			})

		})

		When("overriding with a health-check-type that is not http", func() {
			It("succeeds", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
						PushCommandName, appName,
						"--health-check-type", "port",
						"--no-start",
						"-f", pathToManifest,
					)

					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
