package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with health check type", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		session := helpers.CF("create-app", appName)
		Eventually(session).Should(Exit(0))
	})

	When("setting the app to http health check type", func() {
		It("should update the health check http endpoint", func() {
			helpers.WithMultiEndpointApp(func(dir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "http", "--endpoint", "/third_endpoint.html", "--no-start")).Should(Exit(0))
			})

			session := helpers.CF("get-health-check", appName)
			Eventually(session).Should(Say("web\\s+http\\s+/third_endpoint.html"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("setting the app to port health check type", func() {
		It("should reset the health check http endpoint", func() {
			helpers.WithMultiEndpointApp(func(dir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "port", "--no-start")).Should(Exit(0))
			})

			session := helpers.CF("get-health-check", appName)
			Eventually(session).Should(Say("web\\s+port\\s+1"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("setting the app to process health check type", func() {
		It("should reset the health check http endpoint", func() {
			helpers.WithMultiEndpointApp(func(dir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "process", "--no-start")).Should(Exit(0))
			})

			session := helpers.CF("get-health-check", appName)
			Eventually(session).Should(Say("web\\s+process\\s+1"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("setting the app's health check http endpoint", func() {
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
								"--endpoint", "/third_endpoint.html", "--no-start")
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
											"name":              "some-app",
											"health-check-type": "port",
										},
									},
								})

								session := helpers.CustomCF(
									helpers.CFEnv{WorkingDirectory: dir},
									PushCommandName, appName,
									"--endpoint", "/third_endpoint.html", "--no-start")

								Eventually(session.Err).Should(Say("Incorrect Usage: The flag option --endpoint cannot be used with the manifest property health-check-type set to port"))
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
								"--endpoint", "/third_endpoint.html", "--no-start")
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
})
