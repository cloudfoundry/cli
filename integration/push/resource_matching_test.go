package push

import (
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

	Context("when the app has all of it's resources matched", func() {
		It("does not display the progress bar", func() {
			helpers.WithNoResourceMatchedApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
				Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))

				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-b", "staticfile_buildpack")
				Eventually(session).Should(Say("\\s+name:\\s+%s", appName))
				Eventually(session).Should(Say("All files found in remote cache; nothing to upload."))
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("app", appName)
			Eventually(session).Should(Say("name:\\s+%s", appName))
			Eventually(session).Should(Exit(0))
		})
	})
})
