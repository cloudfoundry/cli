package push

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushing a path with the -p flag", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when the -p and -o flags are used together", func() {
		var path string

		BeforeEach(func() {
			tempFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			path = tempFile.Name()
		})

		AfterEach(func() {
			err := os.Remove(path)
			Expect(err).ToNot(HaveOccurred())
		})

		It("tells the user that they cannot be used together, displays usage and fails", func() {
			session := helpers.CF(PushCommandName, appName, "-o", DockerImage, "-p", path)

			Eventually(session.Err).Should(Say("Incorrect Usage: '--docker-image, -o' and '-p' cannot be used together\\."))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Out).Should(Say("USAGE:"))

			Eventually(session).Should(Exit(1))
		})
	})

	Context("pushing a directory", func() {
		It("pushes the app from the directory", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CF(PushCommandName, appName, "-p", appDir, "-b", "staticfile_buildpack")

				Eventually(session.Out).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session.Out).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session.Out).Should(Say("path:\\s+%s", appDir))
				Eventually(session.Out).Should(Say("routes:"))
				Eventually(session.Out).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session.Out).Should(Say("Packaging files to upload\\.\\.\\."))
				Eventually(session.Out).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session.Out).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				Eventually(session.Out).Should(Say("Staging app and tracing logs\\.\\.\\."))
				Eventually(session.Out).Should(Say("name:\\s+%s", appName))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})
