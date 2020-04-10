package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	//"io"
)

var _ = FDescribe("passwd command", func() {

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
			newPassword		string
			session         *Session
			//passwdInput     io.Reader
			buffer *Buffer
		)
		BeforeEach(func() {
			helpers.LoginCF()
			username = helpers.NewUsername()
			currentPassword = helpers.NewPassword()
			newPassword = helpers.NewPassword()
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

		FWhen("the password is successfully changed", func() {
			It("succeeds", func() {
				buffer = NewBuffer()
				//passwdInput = strings.NewReader(currentPassword+"\n"+newPassword + "\n" + newPassword + "\n")

				//session = helpers.CFWithStdin(passwdInput, "passwd")

				_, err := buffer.Write([]byte(fmt.Sprintf("%s\n", currentPassword)))
				Expect(err).ToNot(HaveOccurred())
				session = helpers.CFWithStdin(buffer, "passwd")
				Eventually(session.Out).Should(Say("New Password>"))
				_, err = buffer.Write([]byte(fmt.Sprintf("%s\n", newPassword)))
				Expect(err).To(Not(HaveOccurred()))
				Eventually(session.Out).Should(Say("Verify Password>"))
				_, err = buffer.Write([]byte(fmt.Sprintf("%s\n", newPassword)))
				Expect(err).To(Not(HaveOccurred()))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("Please log in again"))
				Eventually(session).Should(Exit(0))

				helpers.LoginAs(username, newPassword)
			})
		})

		XWhen("the user enters a different string when verifying their password", func() {
		})

		XWhen("the user enters the wrong current password", func() {})

		XWhen("the user enters the same current and new password", func() {})

		XWhen("the user is not logged in", func() {})
	})
})
