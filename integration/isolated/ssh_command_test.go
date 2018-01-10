package isolated

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

var _ = Describe("ssh command", func() {
	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("ssh", "--help")
				Eventually(session.Out).Should(Say(`NAME:`))
				Eventually(session.Out).Should(Say(`ssh - SSH to an application container instance`))
				Eventually(session.Out).Should(Say(`USAGE:`))
				Eventually(session.Out).Should(Say(`cf ssh APP_NAME \[-i INDEX\] \[-c COMMAND\]\.\.\. \[-L \[BIND_ADDRESS:\]PORT:HOST:HOST_PORT\] \[--skip-host-validation\] \[--skip-remote-execution\] \[--disable-pseudo-tty \| --force-pseudo-tty \| --request-pseudo-tty\]`))
				Eventually(session.Out).Should(Say(`--app-instance-index, -i\s+Application instance index \(Default: 0\)`))
				Eventually(session.Out).Should(Say(`--command, -c\s+Command to run\. This flag can be defined more than once\.`))
				Eventually(session.Out).Should(Say(`--disable-pseudo-tty, -T\s+Disable pseudo-tty allocation`))
				Eventually(session.Out).Should(Say(`--force-pseudo-tty\s+Force pseudo-tty allocation`))
				Eventually(session.Out).Should(Say(`-L\s+Local port forward specification\. This flag can be defined more than once\.`))
				Eventually(session.Out).Should(Say(`--request-pseudo-tty, -t\s+Request pseudo-tty allocation`))
				Eventually(session.Out).Should(Say(`--skip-host-validation, -k\s+Skip host key validation`))
				Eventually(session.Out).Should(Say(`--skip-remote-execution, -N\s+Do not execute a remote command`))
				Eventually(session.Out).Should(Say(`SEE ALSO:`))
				Eventually(session.Out).Should(Say(`allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when an application with multiple instances has been pushed", func() {
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

			setupCF(orgName, spaceName)
			appName = helpers.PrefixedRandomName("app")
			domainName = defaultSharedDomain()
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

		Context("when the app index is specified", func() {
			Context("when it is negative", func() {
				It("throws an error and informs the user that the app instance index cannot be negative", func() {
					session := helpers.CF("ssh", appName, "-i", "-1")
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Out).Should(Say("The application instance index cannot be negative"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the app index exceeds the last valid index", func() {
				It("throws an error and informs the user that the specified application does not exist", func() {
					session := helpers.CF("ssh", appName, "-i", "2")
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Out).Should(Say("The specified application instance does not exist"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when it is a valid index", func() {
				It("does not throw any error", func() {
					buffer := NewBuffer()
					buffer.Write([]byte("exit\n"))
					session := helpers.CFWithStdin(buffer, "ssh", appName, "-i", "0")
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
