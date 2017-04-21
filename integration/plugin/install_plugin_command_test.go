package plugin

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("installing plugins", func() {
	Context("when the plugin contains command names that match core commands", func() {
		It("displays an error on installation", func() {
			session := helpers.CF("install-plugin", "-f", overrideTestPluginPath)
			Eventually(session).Should(Say("Command `push` in the plugin being installed is a native CF command/alias."))
			Eventually(session).Should(Exit(1))
		})
	})
})
