package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unset-env command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("unset-env", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("unset-env - Remove an env variable from an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf unset-env APP_NAME ENV_VAR_NAME"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("ue"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("env, set-env, v3-apps, v3-restart, v3-stage"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
