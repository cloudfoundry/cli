package push

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with health check timeout flag", func() {
	var (
		appName            string
		manifestFilePath   string
		tempDir            string
		healthCheckTimeout = "42"
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		var err error
		tempDir, err = ioutil.TempDir("", "create-manifest")
		Expect(err).ToNot(HaveOccurred())

		manifestFilePath = filepath.Join(tempDir, fmt.Sprintf("%s_manifest.yml", appName))
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	When("the app exists with an http health check", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName, "--health-check-type", "http", "--endpoint", "/",
				)

				Eventually(session).Should(Exit(0))
			})
		})

		It("correctly updates the health check timeout", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName, "--app-start-timeout", healthCheckTimeout,
				)

				Eventually(session).Should(Exit(0))
			})

			session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir},
				"create-app-manifest", appName)
			Eventually(session).Should(Exit(0))

			createdFile, err := ioutil.ReadFile(manifestFilePath)
			Expect(err).ToNot(HaveOccurred())

			Expect(createdFile).To(MatchRegexp("---"))
			Expect(createdFile).To(MatchRegexp("applications:"))
			Expect(createdFile).To(MatchRegexp("name: %s", appName))
			Expect(createdFile).To(MatchRegexp("processes"))
			Expect(createdFile).To(MatchRegexp("timeout: %s", healthCheckTimeout))
		})

	})

})
