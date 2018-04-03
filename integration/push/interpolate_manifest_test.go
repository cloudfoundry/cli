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

var _ = Describe("push with a manifest and a vars file", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when a valid vars file is provided", func() {
		var varsFilePath string

		BeforeEach(func() {
			tmpFile, err := ioutil.TempFile("", "vars-file")
			Expect(err).ToNot(HaveOccurred())
			Expect(tmpFile.Close()).ToNot(HaveOccurred())

			varsFilePath = tmpFile.Name()
			vars := fmt.Sprintf("vars1: %s", appName)
			err = ioutil.WriteFile(varsFilePath, []byte(vars), 0666)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(varsFilePath)).ToNot(HaveOccurred())
		})

		It("pushes the app with the interpolated values in the manifest", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": "((vars1))",
							"path": dir,
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--vars-file", varsFilePath)
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
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
