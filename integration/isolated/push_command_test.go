package isolated

import (
	"fmt"
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

var _ = Describe("Push", func() {
	Context("when the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.PrefixedRandomName("SPACE")

			setupCF(orgName, spaceName)
		})

		Context("when manifest contains non-string env values", func() {
			var appName string

			BeforeEach(func() {
				appName = helpers.PrefixedRandomName("app")
				helpers.WithHelloWorldApp(func(appDir string) {
					manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  env:
    big_float: 123456789.12345678
    big_int: 123412341234
    bool: true
    small_int: 7
    string: "some-string"
`, appName))
					manifestPath := filepath.Join(appDir, "manifest.yml")
					err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
					Expect(err).ToNot(HaveOccurred())

					// Create manifest and add big numbers
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			It("converts all env values to string", func() {
				session := helpers.CF("env", appName)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("big_float: 123456789.12345678"))
				Eventually(session.Out).Should(Say("big_int: 123412341234"))
				Eventually(session.Out).Should(Say("bool: true"))
				Eventually(session.Out).Should(Say("small_int: 7"))
				Eventually(session.Out).Should(Say("string: some-string"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the app has over 260 character paths", func() {
			var tmpDir string

			BeforeEach(func() {
				Skip("Unskip when #134888875 is complete")
				var err error
				tmpDir, err = ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())

				dirName := "dir_name"
				dirNames := []string{}
				for i := 0; i < 32; i++ { // minimum 300 chars, including separators
					dirNames = append(dirNames, dirName)
				}

				fullPath := filepath.Join(tmpDir, filepath.Join(dirNames...))
				if runtime.GOOS == "windows" {
					// `\\?\` is used to skip Windows' file name processor, which imposes
					// length limits. Search MSDN for 'Maximum Path Length Limitation' for
					// more.
					fullPath = `\\?\` + fullPath
				}
				err = os.MkdirAll(fullPath, os.ModeDir|os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(fullPath, "index.html"), []byte("hello world"), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("successfully pushes the app", func() {
				defer os.RemoveAll(tmpDir)
				appName := helpers.PrefixedRandomName("APP")
				session := helpers.CF("push", appName, "-p", tmpDir, "-b", "staticfile_buildpack")
				Eventually(session).Should(Say("1 of 1 instances running"))
				Eventually(session).Should(Say("App %s was started using this command", appName))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when pushing with manifest routes and specifying the -n flag", func() {
			var (
				quotaName     string
				appDir        string
				manifestPath  string
				privateDomain helpers.Domain
				sharedDomain  helpers.Domain
				tcpDomain     helpers.Domain
			)

			BeforeEach(func() {
				quotaName = helpers.PrefixedRandomName("INTEGRATION-QUOTA")

				session := helpers.CF("create-quota", quotaName, "-m", "10G", "-r", "10", "--reserved-route-ports", "4")
				Eventually(session).Should(Exit(0))
				session = helpers.CF("set-quota", orgName, quotaName)
				Eventually(session).Should(Exit(0))

				privateDomain = helpers.NewDomain(orgName, helpers.DomainName("private"))
				privateDomain.Create()
				sharedDomain = helpers.NewDomain(orgName, helpers.DomainName("shared"))
				sharedDomain.CreateShared()
				tcpDomain = helpers.NewDomain(orgName, helpers.DomainName("tcp"))
				tcpDomain.CreateWithRouterGroup("default-tcp")

				var err error
				appDir, err = ioutil.TempDir("", "simple-app")
				Expect(err).ToNot(HaveOccurred())
				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: app-with-routes
  memory: 100M
  instances: 1
  path: .
  routes:
  - route: %s
  - route: %s
  - route: manifest-host.%s/path
  - route: %s:1100
`, privateDomain.Name, sharedDomain.Name, sharedDomain.Name, tcpDomain.Name))
				manifestPath = filepath.Join(appDir, "manifest.yml")
				err = ioutil.WriteFile(manifestPath, manifestContents, 0666)
				Expect(err).ToNot(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(appDir, "index.html"), []byte("hello world"), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should set or replace the route's hostname with the flag value", func() {
				defer os.RemoveAll(appDir)
				var session *Session
				session = helpers.CF("push", helpers.PrefixedRandomName("APP"), "-p", appDir, "-n", "flag-hostname", "-b", "staticfile_buildpack", "-f", manifestPath)

				Eventually(session).Should(Say("Creating route flag-hostname.%s...\nOK", privateDomain.Name))
				Eventually(session).Should(Say("Creating route flag-hostname.%s...\nOK", sharedDomain.Name))
				Eventually(session).Should(Say("Creating route flag-hostname.%s/path...\nOK", sharedDomain.Name))
				Eventually(session).Should(Say("Creating route %s:1100...\nOK", tcpDomain.Name))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
