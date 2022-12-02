package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-user command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("delete-user", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("delete-user - Delete a user"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf delete-user USERNAME \[-f\]`))
				Eventually(session).Should(Say(`cf delete-user USERNAME \[--origin ORIGIN\]`))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("   cf delete-user jsmith                   # internal user"))
				Eventually(session).Should(Say("   cf delete-user jsmith --origin ldap     # LDAP user"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-f\s+Prompt interactively for password`))
				Eventually(session).Should(Say(`--origin\s+Origin for mapping a user account to a user in an external identity provider`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org-users"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "delete-user", "username")
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("one users exist", func() {
			var (
				someUser string
			)

			BeforeEach(func() {
				helpers.LoginCF()
				someUser = helpers.NewUsername()
				somePassword := helpers.NewPassword()
				session := helpers.CF("create-user", someUser, somePassword)
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				session := helpers.CF("delete-user", "-f", someUser)
				Eventually(session).Should(Exit(0))
			})

			It("deletes the user", func() {
				session := helpers.CF("delete-user", someUser, "-f")
				Eventually(session).Should(Say("Deleting user %s...", someUser))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("multiple users exist with the same username", func() {
				BeforeEach(func() {
					session := helpers.CF("create-user", someUser, "--origin", helpers.NonUAAOrigin)
					Eventually(session).Should(Exit(0))
				})

				AfterEach(func() {
					session := helpers.CF("delete-user", "-f", someUser, "--origin", helpers.NonUAAOrigin)
					Eventually(session).Should(Exit(0))
				})

				When("the origin flag is NOT passed in", func() {
					It("errors", func() {
						session := helpers.CF("delete-user", someUser, "-f")
						Eventually(session).Should(Say("Deleting user %s...", someUser))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("The username '%s' is found in multiple origins: (uaa|cli-oidc-provider), (uaa|cli-oidc-provider).", someUser))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the origin flag is passed in", func() {
					It("deletes the correct user", func() {
						session := helpers.CF("delete-user", someUser, "-f", "--origin", helpers.NonUAAOrigin)
						Eventually(session).Should(Say("Deleting user %s...", someUser))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						Expect(helpers.GetUsersV3()).ToNot(ContainElement(helpers.User{
							Username: someUser,
							Origin:   helpers.NonUAAOrigin,
						}))
					})
				})
			})
		})

		When("the user does not exist", func() {
			It("does not error but prints a message letting the user know it never existed", func() {
				session := helpers.CF("delete-user", "nonexistent-user", "-f")
				Eventually(session).Should(Say("User 'nonexistent-user' does not exist."))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
