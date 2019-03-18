package push

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push manifest with a path", func() {
	var (
		appName string

		secondName string
		tempDir    string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		secondName = helpers.NewAppName()
		var err error
		tempDir, err = ioutil.TempDir("", "simple-manifest-test")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
	})

	It("pushes the apps using the path specified", func() {
		helpers.WithHelloWorldApp(func(dir string) {
			manifestPath := filepath.Join(dir, "manifest.yml")
			helpers.WriteManifest(manifestPath, map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": appName,
						"path": dir,
					},
					{
						"name": secondName,
						"path": dir,
					},
				},
			})
			session := helpers.CustomCF(
				helpers.CFEnv{
					EnvVars: map[string]string{"CF_LOG_LEVEL": "debug"},
				},
				PushCommandName,
				appName,
				"-f", manifestPath,
			)

			if runtime.GOOS == "windows" {
				// The paths in windows logging have extra escaping that is difficult
				// to match. Instead match on uploading the right number of files.
				Eventually(session.Err).Should(Say("zipped_file_count=2"))
			} else {
				Eventually(session.Err).Should(helpers.SayPath(`msg="creating archive"\s+Path="?%s"?`, dir))
			}
			Eventually(session).Should(Exit(0))
		})
	})

	When("a single path is not valid", func() {
		It("errors", func() {
			manifestPath := filepath.Join(tempDir, "manifest.yml")
			helpers.WriteManifest(manifestPath, map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": appName,
						"path": "/I/am/a/potato",
					},
					{
						"name": secondName,
						"path": "/baboaboaboaobao/foo",
					},
				},
			})
			session := helpers.CF(PushCommandName, appName, "-f", manifestPath)
			Eventually(session.Err).Should(Say("File not found locally, make sure the file exists at given path /I/am/a/potato"))
			Eventually(session).Should(Exit(1))
		})
	})
})
