package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-env command", func() {
	When("the --help flag provided", func() {
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
	When("the a name and value are provided", func() {
		var (
			orgName   string
			spaceName string
			appName   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.NewAppName()
			helpers.SetupCF(orgName, spaceName)
			helpers.WithEmptyFilesApp(func(appDir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
			})
		})

		It("sets the environment value but doesn't output the value", func() {
			session := helpers.CF("set-env", appName, "key", "value")
			Eventually(session).Should(Exit(0))
			Expect(session).Should(Say("Setting env variable 'key' for app %s in org %s / space %s ", appName, orgName, spaceName))
			Expect(session).Should(Say("OK"))
			session = helpers.CF("restart", appName)
			Eventually(session).Should(Exit(0))
			session = helpers.CF("env", appName)
			Eventually(session).Should(Exit(0))
			Expect(session).Should(Say(`key: value`))
		})
	})
})
