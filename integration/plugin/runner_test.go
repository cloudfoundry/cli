package plugin

import (
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
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
})
