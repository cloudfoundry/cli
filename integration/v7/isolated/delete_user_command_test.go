package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = FDescribe("delete-user command", func() {
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
		var (
			someUser     string
			someLdapUser string
			noobUser     string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			noobUser = helpers.NewUsername()
			noobPassword := helpers.NewPassword()
			session := helpers.CF("create-user", noobUser, noobPassword)
			Eventually(session).Should(Exit(0))

			helpers.LogoutCF()

			env := map[string]string{
				"CF_USERNAME": noobUser,
				"CF_PASSWORD": noobPassword,
			}
			session = helpers.CFWithEnv(env, "auth")
			Eventually(session).Should(Exit(0))
		})

		FWhen("the logged in user is not authorized to delete new users", func() {
			It("fails with insufficient scope error", func() {
				someUser = helpers.NewUsername()
				session := helpers.CF("delete-user", someUser, "-f")
				Eventually(session).Should(Say(`Deleting user %s as %s\.\.\.`, someUser, noobUser))
				Eventually(session.Err).Should(Say(`Server error, status code: 403: Access is denied\.  You do not have privileges to execute this command\.`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the logged in user is authorized to delete new users", func() {
			BeforeEach(func() {
				helpers.LoginCF()
				someUser = helpers.NewUsername()
				somePassword := helpers.NewPassword()
				session := helpers.CF("create-user", someUser, somePassword)
				Eventually(session).Should(Exit(0))
			})

			When("the user to be deleted is found", func() {
				When("the origin flag is NOT passed in", func() {
					When("the user is in uaa", func() {
						It("deletes the user", func() {
							session := helpers.CF("delete-user", someUser, "-f")
							Eventually(session).Should(Say("Deleting user %s...", someUser))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})

						When("the user is NOT in uaa", func() {
							BeforeEach(func() {
								someLdapUser = helpers.NewUsername()
								session := helpers.CF("create-user", someLdapUser, "--origin", "ldap")
								Eventually(session).Should(Exit(0))
							})

							It("errors", func() {
								session := helpers.CF("delete-user", someLdapUser, "-f")
								Eventually(session).Should(Say("User '%s' is found in a different origin, please specify with --origin", someUser)) // TODO: reword
								Eventually(session).Should(Say("FAILED"))
								Eventually(session).Should(Exit(1))
							})
						})
					})

					When("the origin flag is passed in", func() {
						When("the origin is UAA", func() {
							It("deletes the user", func() {
								session := helpers.CF("delete-user", someUser, "-f", "--origin", "uaa")
								Eventually(session).Should(Say("Deleting user %s...", someUser))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("the origin is the empty string", func() {
							When("the user is in uaa", func() {
								It("deletes the user", func() {
									session := helpers.CF("delete-user", someUser, "-f", "--origin", "")
									Eventually(session).Should(Say("Deleting user %s...", someUser))
									Eventually(session).Should(Say("OK"))
									Eventually(session).Should(Exit(0))
								})
							})

						})
					})

					When("the origin is not UAA", func() {
						BeforeEach(func() {
							someLdapUser = helpers.NewUsername()
							session := helpers.CF("create-user", someLdapUser, "--origin", "ldap")
							Eventually(session).Should(Exit(0))
						})

						It("deletes the user", func() {
							session := helpers.CF("delete-user", someUser, "-f", "--origin", "ldap")
							Eventually(session).Should(Say("Deleting user %s...", someUser))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})

					When("argument for flag is not present", func() {
					})
				})
			})

			When("the user does not exist", func() {
				It("errors", func() {
					session := helpers.CF("delete-user", "non-existent-user", "-f")
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("User 'non-existent-user' does not exist."))
					Eventually(session).Should(Exit(0))
				})
			})

			//When("the user does not exist in CC but does exist in UAA", func() {
			//	BeforeEach(func() {
			//		someUser = helpers.NewUsername()
			//		somePassword := helpers.NewPassword()
			//		_, uaaClient, err := shared.NewClients(helpers.GetConfig(), nil, false, "")
			//		_, err = uaaClient.CreateUser(someUser, somePassword, "uaa")
			//		Expect(err).ToNot(HaveOccurred())
			//	})
			//
			//	It("deletes the user", func() {
			//		session := helpers.CF("delete-user", someUser, "-f")
			//		Eventually(session).Should(Say("Deleting user %s...", someUser))
			//		Eventually(session).Should(Say("OK"))
			//		Eventually(session).Should(Exit(0))
			//	})
			//})

			//When("multiple users are found", func() {})
		})
	})
})
