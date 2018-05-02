package plugin

import (
	"os"
	"syscall"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("running plugins", func() {
	Describe("panic handling", func() {
		BeforeEach(func() {
			Eventually(helpers.CF("install-plugin", "-f", panicTestPluginPath)).Should(Exit(0))
		})

		It("will exit 1 if the plugin panics", func() {
			Eventually(helpers.CF("freak-out")).Should(Exit(1))
		})
	})

	Describe("when running plugin commands while CF_HOME is set", func() {
		Context("when CF_PLUGIN_HOME is unset", func() {
			BeforeEach(func() {
				Expect(os.Setenv("CF_PLUGIN_HOME", "")).NotTo(HaveOccurred())
			})

			Context("when a plugin is installed", func() {
				BeforeEach(func() {
					installTestPlugin()
				})

				AfterEach(func() {
					uninstallTestPlugin()
				})

				It("lists the installed plugins", func() {
					session := helpers.CF("plugins")
					Eventually(session).Should(Say("Username"))
					Eventually(session).Should(Exit(0))
				})

				It("is able to run an installed plugin command", func() {
					confirmTestPluginOutput("Username", "admin")
				})
			})
		})

		Context("when CF_PLUGIN_HOME is set", func() {
			Context("when a plugin is installed", func() {
				BeforeEach(func() {
					installTestPlugin()
				})

				AfterEach(func() {
					uninstallTestPlugin()
				})

				It("lists the installed plugins", func() {
					session := helpers.CF("plugins")
					Eventually(session).Should(Say("TestPluginCommandWithAlias"))
					Eventually(session).Should(Exit(0))
				})

				It("can call a command by its alias", func() {
					confirmTestPluginOutput("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", "You called Test Plugin Command With Alias!")
				})
			})
		})
	})

	Describe("signal handling", func() {
		BeforeEach(func() {
			installTestPlugin()
		})

		DescribeTable("will wait for the plugin to terminate",
			func(signal syscall.Signal) {
				session := helpers.CF("Sleep", "3000")
				// Give time for the plugin to be called
				time.Sleep(500 * time.Millisecond)
				session = session.Signal(signal)
				Eventually(session).Should(Say("Slept for 3000 ms"))
				Eventually(session).Should(Exit(0))
			},
			Entry("when SIGINT", syscall.SIGINT),
			Entry("when SIGQUIT", syscall.SIGQUIT),
			Entry("when SIGTERM", syscall.SIGTERM),
		)
	})
})
