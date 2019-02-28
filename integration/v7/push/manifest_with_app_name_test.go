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

var _ = Describe("push with a manifest and an app name", func() {
	var (
		appName        string
		secondName     string
		tempDir        string
		pathToManifest string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		secondName = helpers.NewAppName()
		var err error
		tempDir, err = ioutil.TempDir("", "simple-manifest-test")
		Expect(err).ToNot(HaveOccurred())
		pathToManifest = filepath.Join(tempDir, "manifest.yml")
		helpers.WriteManifest(pathToManifest, map[string]interface{}{
			"applications": []map[string]interface{}{
				{
					"name": appName,
					"env": map[string]interface{}{
						"key1": "val1",
						"key4": false,
					},
				},
				{
					"name": secondName,
					"env": map[string]interface{}{
						"key1": "val1",
						"key4": false,
					},
				},
			},
		})
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
	})

	It("pushes only the specified app", func() {
		helpers.WithHelloWorldApp(func(dir string) {
			session := helpers.CustomCF(
				helpers.CFEnv{WorkingDirectory: dir},
				PushCommandName, appName,
				"-f", pathToManifest,
				"--no-start")
			Eventually(session).Should(Exit(0))
		})

		session := helpers.CF("env", appName)
		Eventually(session).Should(Say(`key1:\s+val1`))
		Eventually(session).Should(Say(`key4:\s+false`))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("env", secondName)
		Eventually(session.Err).Should(Say(fmt.Sprintf("App %s not found", secondName)))
		Eventually(session).Should(Exit(1))
	})

	When("tha app name is not in the manifest", func() {
		It("errors", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(
					helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, "bad-app",
					"-f", pathToManifest,
					"--no-start")
				Eventually(session.Err).Should(Say(`Could not find app named 'bad-app' in manifest`))
				Eventually(session).Should(Exit(1))
			})

		})
	})

})
