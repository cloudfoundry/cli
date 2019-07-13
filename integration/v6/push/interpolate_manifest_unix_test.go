// +build !windows

package push

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with a manifest and vars files via process substitution", func() {
	var (
		appName   string
		instances int

		tmpDir            string
		firstVarsFilePath string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		instances = 4

		var err error
		tmpDir, err = ioutil.TempDir("", "vars-files")
		Expect(err).ToNot(HaveOccurred())

		firstVarsFilePath = filepath.Join(tmpDir, "vars1")
		vars1 := fmt.Sprintf("vars1: %s", appName)
		err = ioutil.WriteFile(firstVarsFilePath, []byte(vars1), 0666)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).ToNot(HaveOccurred())
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

			catFileCmd := fmt.Sprintf("<(cat %s)", firstVarsFilePath)
			command := exec.Command("bash", "-c", fmt.Sprintf("cf %s --vars-file %s --vars-file <(echo vars2: 4)", PushCommandName, catFileCmd))
			command.Dir = dir
			session, err := Start(
				command,
				NewPrefixedWriter(helpers.DebugOutPrefix, GinkgoWriter),
				NewPrefixedWriter(helpers.DebugErrPrefix, GinkgoWriter))
			Expect(err).NotTo(HaveOccurred())
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
