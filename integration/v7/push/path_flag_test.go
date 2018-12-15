// +build !partialPush

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

var _ = When("the -p flag is provided", func() {
	var (
		appName  string
		userName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
		userName, _ = helpers.GetCredentials()
	})

	When("the path is a directory", func() {
		When("the directory contains files", func() {
			It("pushes the app from the directory", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CF(PushCommandName, appName, "-p", appDir)
					Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the directory is empty", func() {
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

	When("the path is a zip file", func() {
		Context("pushing a zip file", func() {
			var archive string

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					tmpfile, err := ioutil.TempFile("", "push-archive-integration")
					Expect(err).ToNot(HaveOccurred())
					archive = tmpfile.Name()
					Expect(tmpfile.Close())

					err = helpers.Zipit(appDir, archive, "")
					Expect(err).ToNot(HaveOccurred())
				})
			})

			AfterEach(func() {
				Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
			})

			It("pushes the app from the zip file", func() {
				session := helpers.CF(PushCommandName, appName, "-p", archive)

				Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the path is a symlink to a directory", func() {
		var symlinkPath string

		BeforeEach(func() {
			tempFile, err := ioutil.TempFile("", "symlink-")
			Expect(err).ToNot(HaveOccurred())
			Expect(tempFile.Close()).To(Succeed())

			symlinkPath = tempFile.Name()
			Expect(os.Remove(symlinkPath)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.Remove(symlinkPath)).To(Succeed())
		})

		It("creates and uploads the package from the directory", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				Expect(os.Symlink(appDir, symlinkPath)).To(Succeed())

				session := helpers.CF(PushCommandName, appName, "-p", symlinkPath)

				Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the path is a symlink to a zip file", func() {
		var (
			archive     string
			symlinkPath string
		)

		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				tmpfile, err := ioutil.TempFile("", "package-archive-integration")
				Expect(err).ToNot(HaveOccurred())
				archive = tmpfile.Name()
				Expect(tmpfile.Close())

				err = helpers.Zipit(appDir, archive, "")
				Expect(err).ToNot(HaveOccurred())
			})

			tempFile, err := ioutil.TempFile("", "symlink-to-archive-")
			Expect(err).ToNot(HaveOccurred())
			Expect(tempFile.Close()).To(Succeed())

			symlinkPath = tempFile.Name()
			Expect(os.Remove(symlinkPath)).To(Succeed())
			Expect(os.Symlink(archive, symlinkPath)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.Remove(archive)).To(Succeed())
			Expect(os.Remove(symlinkPath)).To(Succeed())
		})

		It("creates and uploads the package from the zip file", func() {
			session := helpers.CF(PushCommandName, appName, "-p", symlinkPath)

			Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(`requested state:\s+started`))
			Eventually(session).Should(Exit(0))
		})
	})
})
