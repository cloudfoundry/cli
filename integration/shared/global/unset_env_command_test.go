package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unset-env command", func() {
	When("the --help flag provided", func() {
		It("displays the usage text", func() {
			session := helpers.CF("unset-env", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("unset-env - Remove an env variable"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf unset-env APP_NAME ENV_VAR_NAME"))
			Eventually(session).Should(Say("ALIAS:"))
			Eventually(session).Should(Say("use"))
			Eventually(session).Should(Exit(0))
		})
	})
})
