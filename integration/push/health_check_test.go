package push

import (
	"fmt"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push/update an app using health check type", func() {
	var (
		appName  string
		username string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		username, _ = helpers.GetCredentials()
	})

	Context("updating the application", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "http", "--no-start")
				Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", appName, organization, space, username))
				Eventually(session).Should(Exit(0))

				Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
			})
		})

		Context("when setting the app to http health check type", func() {
			It("should keep the health check http endpoint", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "http", "--no-start")
					Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", appName, organization, space, username))
					Eventually(session).ShouldNot(Say("\\-\\s+health check http endpoint:\\s+/some-endpoint"))
					Eventually(session).ShouldNot(Say("\\+\\s+health check http endpoint:\\s+/"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", helpers.AppGUID(appName)))
				Eventually(session).Should(Say(`"health_check_type":\s+"http"`))
				Eventually(session).Should(Say(`"health_check_http_endpoint":\s+"/"`))
			})
		})

		Context("when setting the app to port health check type", func() {
			It("should reset the health check http endpoint", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "port", "--no-start")
					Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", appName, organization, space, username))
					Eventually(session).Should(Say("\\-\\s+health check http endpoint:\\s+/some-endpoint"))
					Eventually(session).Should(Say("\\-\\s+health check type:\\s+http"))
					Eventually(session).Should(Say("\\+\\s+health check type:\\s+port"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", helpers.AppGUID(appName)))
				Eventually(session).Should(Say(`"health_check_type":\s+"port"`))
				Eventually(session).Should(Say(`"health_check_http_endpoint":\s+""`))
			})
		})

		Context("when setting the app to process health check type", func() {
			It("should reset the health check http endpoint", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "process", "--no-start")
					Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", appName, organization, space, username))
					Eventually(session).Should(Say("\\-\\s+health check http endpoint:\\s+/some-endpoint"))
					Eventually(session).Should(Say("\\-\\s+health check type:\\s+http"))
					Eventually(session).Should(Say("\\+\\s+health check type:\\s+process"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", helpers.AppGUID(appName)))
				Eventually(session).Should(Say(`"health_check_type":\s+"process"`))
				Eventually(session).Should(Say(`"health_check_http_endpoint":\s+""`))
			})
		})
	})

	Context("creating the application", func() {
		Context("when setting a http health check type", func() {
			It("should set the health check type to http and use the default health check endpoint", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "http", "--no-start")

					Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", appName, organization, space, username))
					Eventually(session).Should(Say("\\+\\s+health check http endpoint:\\s+/"))
					Eventually(session).Should(Say("\\+\\s+health check type:\\s+http"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", helpers.AppGUID(appName)))
				Eventually(session).Should(Say(`"health_check_type":\s+"http"`))
				Eventually(session).Should(Say(`"health_check_http_endpoint":\s+"/"`))
			})

			Context("when setting a health check endpoint", func() {
				It("should use the provided endpoint", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name":                       appName,
									"health-check-type":          "http",
									"health-check-http-endpoint": "/some-endpoint",
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
						Eventually(session).Should(Say("Pushing from manifest to org %s / space %s as %s\\.\\.\\.", organization, space, username))
						Eventually(session).Should(Say("\\s+health check http endpoint:\\s+/some-endpoint"))
						Eventually(session).Should(Say("\\s+health check type:\\s+http"))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", helpers.AppGUID(appName)))
					Eventually(session).Should(Say(`"health_check_type":\s+"http"`))
					Eventually(session).Should(Say(`"health_check_http_endpoint":\s+"/some-endpoint"`))
				})
			})
		})

		Context("when setting a port health check type", func() {
			It("it should set the health check type to port", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "port", "--no-start")

					Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", appName, organization, space, username))
					Eventually(session).Should(Say("\\+\\s+health check type:\\s+port"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", helpers.AppGUID(appName)))
				Eventually(session).Should(Say(`"health_check_type":\s+"port"`))
				Eventually(session).Should(Say(`"health_check_http_endpoint":\s+""`))
			})

			Context("when setting an health check endpoint", func() {
				It("should return an error", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name":                       appName,
									"health-check-type":          "port",
									"health-check-http-endpoint": "/some-endpoint",
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
						Eventually(session).Should(Say("Pushing from manifest to org %s / space %s as %s\\.\\.\\.", organization, space, username))
						Eventually(session.Err).Should(Say("Health check type must be 'http' to set a health check HTTP endpoint."))

						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		Context("when setting a process health check type", func() {
			It("it should set the health check type to process", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-u", "process", "--no-start")

					Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", appName, organization, space, username))
					Eventually(session).Should(Say("\\+\\s+health check type:\\s+process"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", helpers.AppGUID(appName)))
				Eventually(session).Should(Say(`"health_check_type":\s+"process"`))
				Eventually(session).Should(Say(`"health_check_http_endpoint":\s+""`))
			})

			Context("when setting a health check endpoint", func() {
				It("should return an error", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name":                       appName,
									"health-check-type":          "process",
									"health-check-http-endpoint": "/some-endpoint",
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
						Eventually(session).Should(Say("Pushing from manifest to org %s / space %s as %s\\.\\.\\.", organization, space, username))
						Eventually(session.Err).Should(Say("Health check type must be 'http' to set a health check HTTP endpoint."))

						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
