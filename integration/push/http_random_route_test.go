package push

import (
	"path"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("HTTP random route", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = "short-app-name" // used on purpose to fit route length requirement
	})

	Context("when passed the --random-route flag", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "--no-start")
				Eventually(session).Should(Say("routes:"))
				Eventually(session).Should(Say("(?i)\\+\\s+%s-[\\w\\d]+-[\\w\\d]+.%s", appName, defaultSharedDomain()))
				Eventually(session).Should(Exit(0))
			})
		})

		It("generates a random route for the app", func() {
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Say("(?i)routes:\\s+%s-[\\w\\d]+-[\\w\\d]+.%s", appName, defaultSharedDomain()))
			Eventually(session).Should(Exit(0))
		})

		Context("when the app has an existing route", func() {
			It("does not generate a random route for the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "--no-start")
					Eventually(session).ShouldNot(Say("\\+\\s+%s-adj", appName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	Context("when passed the random-route manifest property", func() {
		var manifest map[string]interface{}

		BeforeEach(func() {
			manifest = map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name":         appName,
						"random-route": true,
					},
				},
			}

			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(path.Join(dir, "manifest.yml"), manifest)
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
				Eventually(session).Should(Say("routes:"))
				Eventually(session).Should(Say("(?i)\\+\\s+%s-[\\w\\d]+-[\\w\\d]+.%s", appName, defaultSharedDomain()))
				Eventually(session).Should(Exit(0))
			})
		})

		It("generates a random route for the app", func() {
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Say("(?i)routes:\\s+%s-[\\w\\d]+-[\\w\\d]+.%s", appName, defaultSharedDomain()))
			Eventually(session).Should(Exit(0))
		})

		Context("when the app has an existing route", func() {
			It("does not generate a random route for the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(path.Join(dir, "manifest.yml"), manifest)

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
					Eventually(session).ShouldNot(Say("\\+\\s+%s-adj", appName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
