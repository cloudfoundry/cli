package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("triggering legacy push", func() {
	var (
		appName       string
		host          string
		defaultDomain string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		host = helpers.NewAppName()
		defaultDomain = defaultSharedDomain()
	})

	Context("when there are global properties in the manifest", func() {
		It("triggering old push with deprecation warning", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"host": host,
					"applications": []map[string]string{
						{
							"name": appName,
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
				Eventually(session.Err).Should(Say("\\*\\*\\* Global attributes/inheritance in app manifest are not supported in v2-push, delegating to old push \\*\\*\\*"))
				Eventually(session).Should(Say("Creating route %s\\.%s", host, defaultDomain))
			})
		})
	})

	Context("when there is an 'inherit' property in the manifest", func() {
		It("triggering old push with deprecation warning", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "parent.yml"), map[string]interface{}{
					"host": host,
				})

				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"inherit": "./parent.yml",
					"applications": []map[string]string{
						{
							"name": appName,
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
				Eventually(session.Err).Should(Say("\\*\\*\\* Global attributes/inheritance in app manifest are not supported in v2-push, delegating to old push \\*\\*\\*"))
				Eventually(session).Should(Say("Creating route %s\\.%s", host, defaultDomain))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
