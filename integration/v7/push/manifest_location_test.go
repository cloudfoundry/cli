package push

import (
	"io/ioutil"
	"os"
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

		When("the manifest has a .yaml extension", func() {
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

	When("the path to the manifest is specified", func() {
		var workingDir string

		BeforeEach(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "manifest-working-dir")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(workingDir)).ToNot(HaveOccurred())
		})

		When("the manifest has a .yml extension", func() {
			BeforeEach(func() {
				pathToManifest := filepath.Join(workingDir, "manifest.yml")
				helpers.WriteManifest(pathToManifest, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
						},
					},
				})
			})

			It("detects manifests with a .yml suffix", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", workingDir, "--no-start")
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the manifest has a .yaml extension", func() {
			BeforeEach(func() {
				pathToManifest := filepath.Join(workingDir, "manifest.yaml")
				helpers.WriteManifest(pathToManifest, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
						},
					},
				})
			})

			It("detects manifests with a .yaml suffix", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", workingDir, "--no-start")
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the path to the manifest is a directory", func() {
			When("there's no manifest in that directory", func() {
				It("should give a helpful error", func() {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: workingDir}, PushCommandName, "-f", workingDir, "--no-start")
					Eventually(session.Err).Should(Say("Incorrect Usage: The specified directory '%s' does not contain a file named 'manifest.yml'.", workingDir))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
