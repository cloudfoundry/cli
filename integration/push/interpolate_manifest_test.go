package push

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with a manifest and vars files", func() {
	var (
		appName   string
		instances int
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		instances = 4
	})

	Context("when valid vars files are provided", func() {
		var (
			tmpDir string

			firstVarsFilePath  string
			secondVarsFilePath string
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "vars-files")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).ToNot(HaveOccurred())
		})

		Context("when there are no duplicate variables", func() {
			BeforeEach(func() {
				firstVarsFilePath = filepath.Join(tmpDir, "vars1")
				vars1 := fmt.Sprintf("vars1: %s", appName)
				err := ioutil.WriteFile(firstVarsFilePath, []byte(vars1), 0666)
				Expect(err).ToNot(HaveOccurred())

				secondVarsFilePath = filepath.Join(tmpDir, "vars2")
				vars2 := fmt.Sprintf("vars2: %d", instances)
				err = ioutil.WriteFile(secondVarsFilePath, []byte(vars2), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("pushes the app with the interpolated values in the manifest", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":      "((vars1))",
								"instances": "((vars2))",
								"path":      dir,
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--vars-file", firstVarsFilePath, "--vars-file", secondVarsFilePath)
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("\\+\\s+instances:\\s+%d", instances))
					Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Exit(0))
				})
				session := helpers.CF("app", appName)
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when a variable in manifest is not found in var_file", func() {
			BeforeEach(func() {
				firstVarsFilePath = filepath.Join(tmpDir, "vars1")
				vars1 := fmt.Sprintf("vars1: %s", appName)
				err := ioutil.WriteFile(firstVarsFilePath, []byte(vars1), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("fails with error saying that variable is missing", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": "((not_vars))",
								"path": dir,
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--vars-file", firstVarsFilePath)
					Eventually(session.Err).Should(Say("Expected to find variables: not_vars"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when there are duplicate variables", func() {
			BeforeEach(func() {
				firstVarsFilePath = filepath.Join(tmpDir, "vars1")
				vars1 := fmt.Sprintf("vars1: %s\nvars2: %d", "some-garbage-appname", instances)
				err := ioutil.WriteFile(firstVarsFilePath, []byte(vars1), 0666)
				Expect(err).ToNot(HaveOccurred())

				secondVarsFilePath = filepath.Join(tmpDir, "vars2")
				vars2 := fmt.Sprintf("vars1: %s", appName)
				err = ioutil.WriteFile(secondVarsFilePath, []byte(vars2), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("pushes the app using the values from the latter vars-file interpolated in the manifest", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":      "((vars1))",
								"instances": "((vars2))",
								"path":      dir,
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--vars-file", firstVarsFilePath, "--vars-file", secondVarsFilePath)
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("\\+\\s+instances:\\s+%d", instances))
					Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Exit(0))
				})
				session := helpers.CF("app", appName)
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
