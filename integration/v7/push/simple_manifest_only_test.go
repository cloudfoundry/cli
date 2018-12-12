package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with a simple manifest and no flags", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	It("adds or overrides the original env values", func() {
		helpers.WithHelloWorldApp(func(dir string) {
			helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": appName,
						"env": map[string]interface{}{
							"key1": "val1",
							"key4": false,
						},
					},
				},
			})

			session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
			Eventually(session).Should(Exit(0))
		})

		session := helpers.CF("env", appName)
		Eventually(session).Should(Say(`key1:\s+val1`))
		Eventually(session).Should(Say(`key4:\s+false`))
		Eventually(session).Should(Exit(0))
	})
})
