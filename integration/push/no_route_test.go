package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("no-route property", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when pushing with a manifest", func() {
		Context("when pushing a new app", func() {
			It("does not create any routes", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":     appName,
								"no-route": true,
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Consistently(session).ShouldNot(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)routes:\\s+\n"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the app already exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
					Eventually(session).Should(Exit(0))
				})
			})

			It("unmaps any existing routes", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":     appName,
								"no-route": true,
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session).Should(Say("\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("(?i)\\-\\s+%s.%s", appName, defaultSharedDomain()))
					Eventually(session).Should(Say("Unmapping routes\\.\\.\\."))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)routes:\\s+\n"))
				Eventually(session).Should(Exit(0))
			})
		})

		It("does not create any routes", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":     appName,
							"no-route": true,
							"routes": []map[string]string{
								map[string]string{
									"route": "example.com",
								},
							},
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
				Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: no-route, routes", appName))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when pushing with no manifest", func() {
		Context("when pushing a new app", func() {
			It("does not create any routes", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-route", "--no-start")
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Consistently(session).ShouldNot(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)routes:\\s+\n"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the app already exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
					Eventually(session).Should(Exit(0))
				})
			})

			It("unmaps any existing routes", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-route", "--no-start")
					Eventually(session).Should(Say("\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("(?i)\\-\\s+%s.%s", appName, defaultSharedDomain()))
					Eventually(session).Should(Say("Unmapping routes\\.\\.\\."))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)routes:\\s+\n"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when pushing with flags and manifest", func() {
		It("does not create the routes", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"routes": []map[string]string{
								map[string]string{
									"route": "example.com",
								},
							},
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-route", "--no-start")
				Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Say("(?m)routes:\\s+\n"))
			Eventually(session).Should(Exit(0))
		})
	})
})
