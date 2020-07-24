package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-security-group command", func() {
	When("the security-group name is not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-security-group")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SECURITY_GROUP` was not provided"))
			Eventually(session).Should(Say("USAGE"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the security-group does not exist", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("displays a warning and exits 0", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("delete-security-group", "-f", "please-do-not-exist-in-real-life")
			Eventually(session).Should(Say("Deleting security group please-do-not-exist-in-real-life as %s", username))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Say("Security group please-do-not-exist-in-real-life does not exist"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the security group exists", func() {
		var (
			securityGroupName string
			ports             string
			description       string
		)

		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("the -f flag not is provided", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
			})

			When("the user enters 'y'", func() {
				BeforeEach(func() {
					securityGroupName = helpers.NewSecurityGroupName()
					ports = "8080"
					description = "my favorite description"
					securityGroup := helpers.NewSecurityGroup(securityGroupName, "tcp", "0.0.0.0", &ports, &description)
					helpers.CreateSecurityGroup(securityGroup)

					_, err := buffer.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("deletes the security group", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CFWithStdin(buffer, "delete-security-group", securityGroupName)
					Eventually(session).Should(Say("Deleting security group %s as %s", securityGroupName, username))
					Eventually(session).Should(Say(`Really delete the security group %s`, securityGroupName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("security-group", securityGroupName)).Should(Exit(1))
				})
			})

			When("the user enters 'n'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not delete the security group", func() {
					session := helpers.CFWithStdin(buffer, "delete-security-group", securityGroupName)
					Eventually(session).Should(Say(`Really delete the security group %s`, securityGroupName))
					Eventually(session).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the user enters the default input (hits return)", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not delete the security group", func() {
					session := helpers.CFWithStdin(buffer, "delete-security-group", securityGroupName)
					Eventually(session).Should(Say(`Really delete the security group %s`, securityGroupName))
					Eventually(session).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the user enters an invalid answer", func() {
				BeforeEach(func() {
					// The second '\n' is intentional. Otherwise the buffer will be
					// closed while the interaction is still waiting for input; it gets
					// an EOF and causes an error.
					_, err := buffer.Write([]byte("wat\n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("asks again", func() {
					session := helpers.CFWithStdin(buffer, "delete-security-group", securityGroupName)
					Eventually(session).Should(Say(`Really delete the security group %s`, securityGroupName))
					Eventually(session).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the -f flag is provided", func() {
			BeforeEach(func() {
				securityGroupName = helpers.NewSecurityGroupName()
				ports = "8080"
				description = "my favorite description"
				securityGroup := helpers.NewSecurityGroup(securityGroupName, "tcp", "0.0.0.0", &ports, &description)
				helpers.CreateSecurityGroup(securityGroup)
			})

			It("deletes the security group", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("delete-security-group", securityGroupName, "-f")
				Eventually(session).Should(Say("Deleting security group %s as %s", securityGroupName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				Eventually(helpers.CF("security-group", securityGroupName)).Should(Exit(1))
			})
		})
	})
})
