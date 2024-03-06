package push

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Push with manifest variable interpolation", func() {
	var (
		appName   string
		instances int

		manifestPath string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		instances = 4

		tmp, err := ioutil.TempFile("", "manifest-interpolation")
		Expect(err).ToNot(HaveOccurred())
		Expect(tmp.Close()).ToNot(HaveOccurred())
		manifestPath = tmp.Name()

		helpers.WriteManifest(manifestPath, map[string]interface{}{
			"applications": []map[string]interface{}{
				{
					"name":      "((vars1))",
					"instances": "((vars2))",
				},
			},
		})
	})

	AfterEach(func() {
		Expect(os.RemoveAll(manifestPath)).ToNot(HaveOccurred())
	})

	When("only `--vars-file` flags are provided", func() {
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

		When("there are no duplicate variables", func() {
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

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", manifestPath, "--vars-file", firstVarsFilePath, "--vars-file", secondVarsFilePath)
					Eventually(session).Should(Say(`Getting app info\.\.\.`))
					Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
					Eventually(session).Should(Say(`\+\s+name:\s+%s`, appName))
					Eventually(session).Should(Say(`\+\s+instances:\s+%d`, instances))
					Eventually(session).Should(Say(`Mapping routes\.\.\.`))
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Exit(0))
				})
				session := helpers.CF("app", appName)
				Eventually(session).Should(Exit(0))
			})
		})

		When("a variable in manifest is not found in var_file", func() {
			BeforeEach(func() {
				firstVarsFilePath = filepath.Join(tmpDir, "vars1")
				vars1 := fmt.Sprintf("vars1: %s", appName)
				err := ioutil.WriteFile(firstVarsFilePath, []byte(vars1), 0666)
				Expect(err).ToNot(HaveOccurred())

				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": "((not_vars))",
						},
					},
				})
			})

			It("fails with error saying that variable is missing", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", manifestPath, "--vars-file", firstVarsFilePath)
					Eventually(session.Err).Should(Say("Expected to find variables: not_vars"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("there are duplicate variables", func() {
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
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", manifestPath, "--vars-file", firstVarsFilePath, "--vars-file", secondVarsFilePath)
					Eventually(session).Should(Say(`Getting app info\.\.\.`))
					Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
					Eventually(session).Should(Say(`\+\s+name:\s+%s`, appName))
					Eventually(session).Should(Say(`\+\s+instances:\s+%d`, instances))
					Eventually(session).Should(Say(`Mapping routes\.\.\.`))
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Exit(0))
				})
				session := helpers.CF("app", appName)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("only `--var` flag vars are provided", func() {
		It("replaces the variables with the provided values", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", manifestPath, "--var", fmt.Sprintf("vars1=%s", appName), "--var", fmt.Sprintf("vars2=%d", instances))
				Eventually(session).Should(Say(`Getting app info\.\.\.`))
				Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
				Eventually(session).Should(Say(`\+\s+name:\s+%s`, appName))
				Eventually(session).Should(Say(`\+\s+instances:\s+%d`, instances))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Exit(0))
			})
			session := helpers.CF("app", appName)
			Eventually(session).Should(Exit(0))
		})
	})

	When("`--vars-file` and `--var` flag vars are provided", func() {
		var varsFilePath string
		BeforeEach(func() {
			tmp, err := ioutil.TempFile("", "varsfile-interpolation")
			Expect(err).ToNot(HaveOccurred())
			Expect(tmp.Close()).ToNot(HaveOccurred())

			varsFilePath = tmp.Name()
			vars1 := fmt.Sprintf("vars1: %s\nvars2: %d", "some-garbage-appname", instances)
			Expect(ioutil.WriteFile(varsFilePath, []byte(vars1), 0666)).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(varsFilePath)).ToNot(HaveOccurred())
		})

		It("overwrites the vars-file with the provided var key value pair", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "-f", manifestPath, "--var", fmt.Sprintf("vars1=%s", appName), "--vars-file", varsFilePath)
				Eventually(session).Should(Say(`Getting app info\.\.\.`))
				Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
				Eventually(session).Should(Say(`\+\s+name:\s+%s`, appName))
				Eventually(session).Should(Say(`\+\s+instances:\s+%d`, instances))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Exit(0))
			})
			session := helpers.CF("app", appName)
			Eventually(session).Should(Exit(0))
		})
	})
})
