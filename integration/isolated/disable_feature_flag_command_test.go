package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("disable-feature-flag command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("disable-feature-flag", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("disable-feature-flag - Prevent use of a feature"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf disable-feature-flag FEATURE_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("enable-feature-flag, feature-flags"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
