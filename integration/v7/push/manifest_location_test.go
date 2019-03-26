package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("reading of the manifest based on location", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
	})

	When("the manifest exists in the current directory", func() {
		When("the manifest has a .yml extension", func() {
			It("detects manifests with a .yml suffix", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					pathToManifest := filepath.Join(dir, "manifest.yml")
					helpers.WriteManifest(pathToManifest, map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": appName,
							},
						},
					})
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the manifest has a .yml extension", func() {
			It("detects manifests with a .yaml suffix", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					pathToManifest := filepath.Join(dir, "manifest.yaml")
					helpers.WriteManifest(pathToManifest, map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": appName,
							},
						},
					})
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
