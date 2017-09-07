package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-env command", func() {
	Context("when the --help flag provided", func() {
		It("displays the usage text", func() {
			session := helpers.CF("set-env", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("set-env - Set an env variable for an app"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE"))
			Eventually(session).Should(Say("ALIAS:"))
			Eventually(session).Should(Say("se"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("apps, env, restart, set-running-environment-variable-group, set-staging-environment-variable-group, unset-env"))
			Eventually(session).Should(Exit(0))
		})
	})
})
