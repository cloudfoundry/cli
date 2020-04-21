package isolated

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("ssh command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	Describe("help", func() {
		When("--help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("ssh", "--help")
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`ssh - SSH to an application container instance`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`cf ssh APP_NAME \[-i INDEX\] \[-c COMMAND\]\.\.\. \[-L \[BIND_ADDRESS:\]PORT:HOST:HOST_PORT\] \[--skip-host-validation\] \[--skip-remote-execution\] \[--disable-pseudo-tty \| --force-pseudo-tty \| --request-pseudo-tty\]`))
				Eventually(session).Should(Say(`--app-instance-index, -i\s+Application instance index \(Default: 0\)`))
				Eventually(session).Should(Say(`--command, -c\s+Command to run\. This flag can be defined more than once\.`))
				Eventually(session).Should(Say(`--disable-pseudo-tty, -T\s+Disable pseudo-tty allocation`))
				Eventually(session).Should(Say(`--force-pseudo-tty\s+Force pseudo-tty allocation`))
				Eventually(session).Should(Say(`-L\s+Local port forward specification\. This flag can be defined more than once\.`))
				Eventually(session).Should(Say(`--request-pseudo-tty, -t\s+Request pseudo-tty allocation`))
				Eventually(session).Should(Say(`--skip-host-validation, -k\s+Skip host key validation`))
				Eventually(session).Should(Say(`--skip-remote-execution, -N\s+Do not execute a remote command`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("an application with multiple instances has been pushed", func() {
		var (
			appName          string
			appDirForCleanup string
			domainName       string
			orgName          string
			spaceName        string
		)

		BeforeEach(func() {
			helpers.LogoutCF()

			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
			appName = helpers.PrefixedRandomName("app")
			domainName = helpers.DefaultSharedDomain()
			helpers.WithHelloWorldApp(func(appDir string) {
				manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  instances: 2
  disk_quota: 128M
  routes:
  - route: %s.%s
`, appName, appName, domainName))
				manifestPath := filepath.Join(appDir, "manifest.yml")
				err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
				Expect(err).ToNot(HaveOccurred())

				// Create manifest
				Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack", "--random-route")).Should(Exit(0))
				appDirForCleanup = appDir
			})
		})

		AfterEach(func() {
			Eventually(helpers.CF("delete", appName, "-f", "-r")).Should(Exit(0))
			Expect(os.RemoveAll(appDirForCleanup)).NotTo(HaveOccurred())
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app index is specified", func() {
			When("it is negative", func() {
				It("throws an error and informs the user that the app instance index cannot be negative", func() {
					session := helpers.CF("ssh", appName, "-i", "-1")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("The application instance index cannot be negative"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the app index exceeds the last valid index", func() {
				It("throws an error and informs the user that the specified application does not exist", func() {
					session := helpers.CF("ssh", appName, "-i", "2")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("The specified application instance does not exist"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("it is a valid index", func() {
				It("does not throw any error", func() {
					buffer := NewBuffer()
					_, err := buffer.Write([]byte("exit\n"))
					Expect(err).NotTo(HaveOccurred())

					By("waiting for all instances to be running")
					Eventually(func() bool {
						session := helpers.CF("app", appName)
						Eventually(session).Should(Exit(0))
						output := session.Out.Contents()
						return regexp.MustCompile(`#\d.*running.*\n#\d.*running.*`).Match(output)
					}, 30*time.Second, 2*time.Second).Should(BeTrue())

					session := helpers.CFWithStdin(buffer, "ssh", appName, "-i", "0")
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
