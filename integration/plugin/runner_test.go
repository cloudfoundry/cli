package plugin

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("running plugins", func() {
	Describe("plugin command alias", func() {
		BeforeEach(func() {
			installTestPlugin()
		})

		AfterEach(func() {
			uninstallTestPlugin()
		})

		It("can call a command by its alias", func() {
			confirmTestPluginOutput("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", "You called Test Plugin Command With Alias!")
		})
	})

	Describe("panic handling", func() {
		BeforeEach(func() {
			session := helpers.CF("install-plugin", "-f", panicTestPluginPath)
			Eventually(session).Should(Exit(0))
		})

		It("will exit 1 if the plugin panics", func() {
			session := helpers.CF("freak-out")
			Eventually(session).Should(Exit(1))
		})
	})
})
