package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-user command", func() {
	When("the logged in user is authorized to delete-users", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("deleting a user that exists in multiple origins", func() {
			var newUser string

			BeforeEach(func() {
				newUser = helpers.NewUsername()

				Eventually(helpers.CF("create-user", newUser, "--origin", "ldap")).Should(Exit(0))
				Eventually(helpers.CF("create-user", newUser, helpers.NewPassword())).Should(Exit(0))
			})

			AfterEach(func() {
				// Doing the cleanup here because it can't easily be done in
				// bin/cleanup-integration.
				users := helpers.GetUsers()

				var usersDeleted int
				for _, user := range users {
					if user.Username == newUser {
						Eventually(helpers.CF("curl", "-X", "DELETE", fmt.Sprintf("/v2/users/%s", user.GUID))).Should(Exit(0))
						usersDeleted++
					}
				}

				Expect(usersDeleted).To(Equal(2), "some users were not deleted")
			})

			It("errors with DuplicateUsernameError", func() {
				session := helpers.CF("delete-user", "-f", newUser)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Error deleting user %s", newUser))
				Eventually(session).Should(Say("The user exists in multiple origins."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
