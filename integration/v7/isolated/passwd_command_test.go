package isolated

import (
	"fmt"
	"io"
	"strings"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("passwd command", func() {

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("passwd", "GETTING STARTED", "Change user password"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("passwd", "--help")

				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say("passwd - Change user password"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf passwd"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("pw"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("functionality", func() {
		var (
			username        string
			currentPassword string
			newPassword     string
			verifyPassword  string
			session         *Session
			passwdInput     io.Reader
		)

		BeforeEach(func() {
			helpers.LoginCF()
			username = helpers.NewUsername()
			currentPassword = helpers.NewPassword()
			newPassword = helpers.NewPassword()
			verifyPassword = newPassword
			session = helpers.CF("create-user", username, currentPassword)
			Eventually(session).Should(Exit(0))
			helpers.LogoutCF()

			env := map[string]string{
				"CF_USERNAME": username,
				"CF_PASSWORD": currentPassword,
			}
			session = helpers.CFWithEnv(env, "auth")
			Eventually(session).Should(Exit(0))
		})

		AfterEach(func() {
			helpers.LoginCF()
			helpers.DeleteUser(username)
		})

		JustBeforeEach(func() {
			passwdInput = strings.NewReader(fmt.Sprintf(`%s
%s
%s
`, currentPassword, newPassword, verifyPassword))
			session = helpers.CFWithStdin(passwdInput, "passwd")
		})

		When("the password is successfully changed", func() {
			It("succeeds", func() {
				Eventually(session.Out).Should(Say("Current Password>"))
				Eventually(session.Out).Should(Say("New Password>"))
				Eventually(session.Out).Should(Say("Verify Password>"))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("Please log in again"))
				Eventually(session).Should(Exit(0))

				helpers.LoginAs(username, newPassword)
			})
		})

		When("the user's new passwords don't match", func() {
			BeforeEach(func() {
				verifyPassword = "passwood-dont-match"
			})

			It("fails with an error message", func() {
				Eventually(session.Out).Should(Say("Current Password>"))
				Eventually(session.Out).Should(Say("New Password>"))
				Eventually(session.Out).Should(Say("Verify Password>"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Password verification does not match"))
				Eventually(session).Should(Exit(1))

				helpers.LoginAs(username, currentPassword)
			})
		})

		When("the user enters the wrong current password", func() {
			BeforeEach(func() {
				currentPassword = "definitely not the right password"
			})

			It("fails with an error message", func() {
				Eventually(session.Out).Should(Say("Current Password>"))
				Eventually(session.Out).Should(Say("New Password>"))
				Eventually(session.Out).Should(Say("Verify Password>"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Current password did not match"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the user enters the same current and new password", func() {
			BeforeEach(func() {
				newPassword = currentPassword
				verifyPassword = currentPassword
			})

			It("fails with an error message", func() {
				Eventually(session.Out).Should(Say("Current Password>"))
				Eventually(session.Out).Should(Say("New Password>"))
				Eventually(session.Out).Should(Say("Verify Password>"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Server error"))
				Eventually(session).Should(Exit(1))

				helpers.LoginAs(username, currentPassword)
			})
		})

		When("the user enters a zero-length new password", func() {
			BeforeEach(func() {
				newPassword = ""
				verifyPassword = ""
			})

			It("fails with an error message", func() {
				Eventually(session.Out).Should(Say("Current Password>"))
				Eventually(session.Out).Should(Say("New Password>"))
				Eventually(session.Out).Should(Say("Verify Password>"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Server error"))
				Eventually(session).Should(Exit(1))

				helpers.LoginAs(username, currentPassword)
			})
		})

		When("the user is not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with an error message", func() {
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
