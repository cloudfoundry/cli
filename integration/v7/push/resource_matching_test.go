package push

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("resource matching", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	When("the app has some of it's resources matched", func() {
		It("uploads the unmatched resources", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				tempfile := filepath.Join(dir, "ignore.html")
				err := ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %s %s", time.Now(), strings.Repeat("a", 65*1024))), 0666)
				Expect(err).ToNot(HaveOccurred())

				session := helpers.DebugCustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
				Eventually(session.Err).Should(Say("zipped_file_count=3"))
				Eventually(session).Should(Exit(0))

				session = helpers.DebugCustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-b", "staticfile_buildpack")
				Eventually(session.Err).Should(Say("zipped_file_count=2"))
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("app", appName)
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the app has all of it's resources matched", func() {
		It("does not display the progress bar", func() {
			// Skip("until #164837999 is complete")
			helpers.WithNoResourceMatchedApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
				Eventually(session).Should(Say(`\s+name:\s+%s`, appName))
				Eventually(session).Should(Exit(0))

				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-b", "staticfile_buildpack")
				Eventually(session).Should(Say("All files found in remote cache; nothing to upload."))
				Eventually(session).Should(Say(`\s+name:\s+%s`, appName))
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("app", appName)
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the app has only empty files", func() {
		It("skips resource matching", func() {
			helpers.WithEmptyFilesApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
				Eventually(session).Should(Say(`\s+name:\s+%s`, appName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
