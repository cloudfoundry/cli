package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-user command", func() {
	BeforeEach(func() {
		helpers.RunIfExperimental("")
	})
	Context("when the logged in user is authorized to delete-users", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})
		Context("when deleting a user that exists in multiple origins", func() {
			var (
				newUser string
			)
			BeforeEach(func() {
				newUser = helpers.RandomUsername()
				session := helpers.CF("create-user", newUser, "--origin", "ldap")
				Eventually(session).Should(Exit(0))
				session = helpers.CF("create-user", newUser, helpers.RandomPassword())
				Eventually(session).Should(Exit(0))
				session = helpers.CF("create-user", newUser, "--origin", "some-other-origin")
				Eventually(session).Should(Exit(0))
			})
			It("errors with DuplicateUsernameError", func() {
				session := helpers.CF("delete-user", "-f", newUser)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Error deleting user %s.", newUser))
				Eventually(session.Out).Should(Say("Multiple users with that username returned. Please use 'cf curl' with specific origin instead."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
