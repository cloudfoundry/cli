package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("staging-environment-variable-group command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("staging-environment-variable-group", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("staging-environment-variable-group - Retrieve the contents of the staging environment variable group"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf staging-environment-variable-group"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("sevg"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("env, running-environment-variable-group"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "staging-environment-variable-group")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.LoginCF()
			userName, _ = helpers.GetCredentials()
		})

		When("there are no variables in the group", func() {
			It("displays an empty-list message", func() {
				session := helpers.CF("staging-environment-variable-group")

				Eventually(session).Should(Say(`Getting the staging environment variable group as %s\.\.\.`, userName))
				Eventually(session).Should(Say(`No staging environment variable group has been set\.`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("there are variables in the group", func() {
			BeforeEach(func() {
				envVars := `{"key_one": "one"}`
				session := helpers.CF("set-staging-environment-variable-group", envVars)
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				envVars := `{}`
				session := helpers.CF("set-staging-environment-variable-group", envVars)
				Eventually(session).Should(Exit(0))
			})

			It("displays the environment variables in a table", func() {
				session := helpers.CF("staging-environment-variable-group")

				Eventually(session).Should(Say(`Getting the staging environment variable group as %s\.\.\.`, userName))
				Eventually(session).Should(Say(`variable name\s+assigned value`))
				Eventually(session).Should(Say(`key_one\s+one`))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
