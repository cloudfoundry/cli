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
		appName                  string
		randomHostName           string
		tempDir                  string
		pathToManifestForNoRoute string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		domainName := helpers.DefaultSharedDomain()
		randomHostName = helpers.RandomName()

		var err error
		tempDir, err = ioutil.TempDir("", "no-route-flag-with-manifest-test")
		Expect(err).ToNot(HaveOccurred())
		pathToSetupManifest := filepath.Join(tempDir, "setup-manifest.yml")
		helpers.WriteManifest(pathToSetupManifest, map[string]interface{}{
			"applications": []map[string]interface{}{
				{
					"name": appName,
					"routes": []map[string]string{
						{"route": fmt.Sprintf("%s.%s", appName, domainName)},
					},
				},
			},
		})

		helpers.WithHelloWorldApp(func(dir string) {
			session := helpers.CustomCF(
				helpers.CFEnv{WorkingDirectory: dir},
				PushCommandName, appName,
				"-f", pathToSetupManifest,
				"--no-start")
			Eventually(session).Should(Exit(0))
		})

		pathToManifestForNoRoute = filepath.Join(tempDir, "no-route-manifest.yml")
		helpers.WriteManifest(pathToManifestForNoRoute, map[string]interface{}{
			"applications": []map[string]interface{}{
				{
					"name": appName,
					"routes": []map[string]string{
						{"route": fmt.Sprintf("%s.%s", randomHostName, domainName)},
					},
				},
			},
		})
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
	})

	It("pushes the app, doesnt map new routes, and removes old routes", func() {
		helpers.WithHelloWorldApp(func(dir string) {
			session := helpers.CustomCF(
				helpers.CFEnv{WorkingDirectory: dir},
				PushCommandName, appName,
				"-f", pathToManifestForNoRoute,
				"--no-start", "--no-route")
			Eventually(session).Should(Exit(0))
		})

		session := helpers.CF("app", appName)
		Eventually(session).Should(Say(`(?m)routes:\s+$`))
		Eventually(session).Should(Exit(0))

	})

})
