package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("no-route", func() {

	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
	})

	When("the --no-route flag is set", func() {
		It("does not map any routes to the app", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName, "--no-start", "--no-route")
				Consistently(session).ShouldNot(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+stopped`))
				Eventually(session).Should(Say(`(?m)routes:\s+\n`))
				Eventually(session).Should(Exit(0))
			})
		})

		It("unmaps currently mapped routes", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName, "--no-start")
				Eventually(session).Should(Exit(0))

				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName, "--no-start", "--no-route")
				Consistently(session).ShouldNot(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+stopped`))
				Eventually(session).Should(Say(`(?m)routes:\s+\n`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the no-route is in the manifest", func() {
		It("does not map any routes to the app", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				manifestPath := filepath.Join(appDir, "manifest.yml")
				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":     appName,
							"no-route": true,
						},
					},
				})
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName, "--no-start")
				Consistently(session).ShouldNot(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+stopped`))
				Eventually(session).Should(Say(`(?m)routes:\s+\n`))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
