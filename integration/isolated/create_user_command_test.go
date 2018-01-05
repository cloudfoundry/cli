package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-user command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("create-user", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("create-user - Create a new user"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf create-user USERNAME PASSWORD"))
				Eventually(session.Out).Should(Say("cf create-user USERNAME --origin ORIGIN"))
				Eventually(session.Out).Should(Say("EXAMPLES:"))
				Eventually(session.Out).Should(Say("   cf create-user j.smith@example.com S3cr3t                  # internal user"))
				Eventually(session.Out).Should(Say("   cf create-user j.smith@example.com --origin ldap           # LDAP user"))
				Eventually(session.Out).Should(Say("   cf create-user j.smith@example.com --origin provider-alias # SAML or OpenID Connect federated user"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("--origin      Origin for mapping a user account to a user in an external identity provider"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("passwd, set-org-role, set-space-role"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-user", "username", "password")
		})
	})

	Context("when the environment is setup correctly", func() {
		Context("when the logged in user is not authorized to create new users", func() {
			var (
				newUser     string
				newPassword string
			)

			BeforeEach(func() {
				helpers.LoginCF()
				noobUser := helpers.NewUsername()
				noobPassword := helpers.NewPassword()
				session := helpers.CF("create-user", noobUser, noobPassword)
				Eventually(session).Should(Exit(0))
				session = helpers.CF("auth", noobUser, noobPassword)
				Eventually(session).Should(Exit(0))
				newUser = helpers.NewUsername()
				newPassword = helpers.NewPassword()
			})

			It("fails with insufficient scope error", func() {
				session := helpers.CF("create-user", newUser, newPassword)
				Eventually(session.Out).Should(Say("Creating user %s\\.\\.\\.", newUser))
				Eventually(session.Out).Should(Say("Error creating user %s\\.", newUser))
				Eventually(session.Err).Should(Say("You are not authorized to perform the requested action\\."))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the logged in user is authorized to create new users", func() {
			BeforeEach(func() {
				helpers.LoginCF()
			})

			Context("when passed invalid username", func() {
				DescribeTable("when passed funkyUsername",
					func(funkyUsername string) {
						session := helpers.CF("create-user", funkyUsername, helpers.NewPassword())
						Eventually(session.Out).Should(Say("Error creating user %s.", funkyUsername))
						Eventually(session.Err).Should(Say("Username must match pattern: \\[\\\\p\\{L\\}\\+0\\-9\\+\\\\\\-_\\.@'!\\]\\+"))
						Eventually(session.Out).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					},

					Entry("fails when passed an emoji", "😀"),
					Entry("fails when passed a backtick", "`"),
				)

				Context("when the username is empty", func() {
					It("fails with a username must be provided error", func() {
						session := helpers.CF("create-user", "", helpers.NewPassword())
						Eventually(session.Out).Should(Say("Error creating user ."))
						Eventually(session.Err).Should(Say("A username must be provided."))
						Eventually(session.Out).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the user passes in an origin flag", func() {
				Context("when the origin is UAA", func() {
					Context("when password is not present", func() {
						It("errors and prints usage", func() {
							newUser := helpers.NewUsername()
							session := helpers.CF("create-user", newUser, "--origin", "UAA")
							Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `PASSWORD` was not provided"))
							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Out).Should(Say("USAGE"))
							Eventually(session).Should(Exit(1))
						})
					})
				})
				Context("when the origin is the empty string", func() {
					Context("when password is not present", func() {
						It("errors and prints usage", func() {
							newUser := helpers.NewUsername()
							session := helpers.CF("create-user", newUser, "--origin", "")
							Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `PASSWORD` was not provided"))
							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Out).Should(Say("USAGE"))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("when the origin is not UAA or empty", func() {
					It("creates the new user in the specified origin", func() {
						newUser := helpers.NewUsername()
						session := helpers.CF("create-user", newUser, "--origin", "ldap")
						Eventually(session.Out).Should(Say("Creating user %s...", newUser))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("TIP: Assign roles with 'cf set-org-role' and 'cf set-space-role'"))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when argument for flag is not present", func() {
					It("fails with incorrect usage error", func() {
						session := helpers.CF("create-user", helpers.NewUsername(), "--origin")
						Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `--origin'"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when password is not present", func() {
				It("fails with incorrect usage error", func() {
					session := helpers.CF("create-user", helpers.NewUsername())
					Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `PASSWORD` was not provided"))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Out).Should(Say("USAGE"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the user already exists", func() {
				var (
					newUser     string
					newPassword string
				)

				BeforeEach(func() {
					newUser = helpers.NewUsername()
					newPassword = helpers.NewPassword()
					session := helpers.CF("create-user", newUser, newPassword)
					Eventually(session).Should(Exit(0))
				})

				It("fails with the user already exists message", func() {
					session := helpers.CF("create-user", newUser, newPassword)
					Eventually(session.Err).Should(Say("user %s already exists", newUser))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the user does not already exist", func() {
				It("creates the new user", func() {
					newUser := helpers.NewUsername()
					newPassword := helpers.NewPassword()
					session := helpers.CF("create-user", newUser, newPassword)
					Eventually(session.Out).Should(Say("Creating user %s...", newUser))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("TIP: Assign roles with 'cf set-org-role' and 'cf set-space-role'"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
