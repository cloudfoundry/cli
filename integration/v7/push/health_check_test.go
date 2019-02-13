package push

import (
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
