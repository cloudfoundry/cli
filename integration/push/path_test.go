package push

import (
	"io/ioutil"
	"os"
	"regexp"

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

	Context("pushing a directory", func() {
		Context("when the directory contains files", func() {
			It("pushes the app from the directory", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CF(PushCommandName, appName, "-p", appDir)

					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("path:\\s+%s", regexp.QuoteMeta(appDir)))
					Eventually(session).Should(Say("routes:"))
					Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Say("Comparing local files to remote cache\\.\\.\\."))
					Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
					Eventually(session).Should(Say("Uploading files\\.\\.\\."))
					Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
					Eventually(session).Should(Say("Staging app and tracing logs\\.\\.\\."))
					Eventually(session).Should(Say("name:\\s+%s", appName))

					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the directory is empty", func() {
			var emptyDir string

			BeforeEach(func() {
				var err error
				emptyDir, err = ioutil.TempDir("", "integration-push-path-empty")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.RemoveAll(emptyDir)).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				session := helpers.CF(PushCommandName, appName, "-p", emptyDir)
				Eventually(session.Err).Should(Say("No app files found in '%s'", regexp.QuoteMeta(emptyDir)))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("pushing a zip file", func() {
		var archive string

		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				tmpfile, err := ioutil.TempFile("", "push-archive-integration")
				Expect(err).ToNot(HaveOccurred())
				archive = tmpfile.Name()
				Expect(tmpfile.Close()).ToNot(HaveOccurred())

				err = helpers.Zipit(appDir, archive, "")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		AfterEach(func() {
			Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
		})

		It("pushes the app from the zip file", func() {
			session := helpers.CF(PushCommandName, appName, "-p", archive)

			Eventually(session).Should(Say("Getting app info\\.\\.\\."))
			Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
			Eventually(session).Should(Say("path:\\s+%s", regexp.QuoteMeta(archive)))
			Eventually(session).Should(Say("routes:"))
			Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
			Eventually(session).Should(Say("Comparing local files to remote cache\\.\\.\\."))
			Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
			Eventually(session).Should(Say("Uploading files\\.\\.\\."))
			Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
			Eventually(session).Should(Say("Staging app and tracing logs\\.\\.\\."))
			Eventually(session).Should(Say("name:\\s+%s", appName))

			Eventually(session).Should(Exit(0))
		})
	})
})
