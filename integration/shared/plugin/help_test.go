package plugin

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("help", func() {
	BeforeEach(func() {
		installTestPlugin()
	})

	AfterEach(func() {
		uninstallTestPlugin()
	})

	It("displays the plugin commands in master help", func() {
		session := helpers.CF("help")
		Eventually(session).Should(Say("TestPluginCommandWithAlias"))
		Eventually(session).Should(Exit(0))
	})

	DescribeTable("displays individual plugin help",
		func(helpCommand ...string) {
			session := helpers.CF(helpCommand...)
			Eventually(session).Should(Say("TestPluginCommandWithAlias"))
			Eventually(session).Should(Say("This is my plugin help test. Banana."))
			Eventually(session).Should(Say("I R Usage"))
			Eventually(session).Should(Say("--dis-flag\\s+is a flag"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when passed to help", "help", "TestPluginCommandWithAlias"),
		Entry("when when passed -h", "TestPluginCommandWithAlias", "-h"),
		Entry("when when passed --help", "TestPluginCommandWithAlias", "--help"),
	)
})
