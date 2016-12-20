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
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("NAME:"))
				Expect(session.Out).To(Say("create-user - Create a new user"))
				Expect(session.Out).To(Say("USAGE:"))
				Expect(session.Out).To(Say("cf create-user USERNAME PASSWORD"))
				Expect(session.Out).To(Say("cf create-user USERNAME --origin ORIGIN"))
				Expect(session.Out).To(Say("EXAMPLES:"))
				Expect(session.Out).To(Say("   cf create-user j.smith@example.com S3cr3t                  # internal user"))
				Expect(session.Out).To(Say("   cf create-user j.smith@example.com --origin ldap           # LDAP user"))
				Expect(session.Out).To(Say("   cf create-user j.smith@example.com --origin provider-alias # SAML or OpenID Connect federated user"))
				Expect(session.Out).To(Say("OPTIONS:"))
				Expect(session.Out).To(Say("--origin      Origin for mapping a user account to a user in an external identity provider"))
				Expect(session.Out).To(Say("SEE ALSO:"))
				Expect(session.Out).To(Say("passwd, set-org-role, set-space-role"))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("create-user", helpers.RandomUsername(), helpers.RandomPassword())
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("create-user", helpers.RandomUsername(), helpers.RandomPassword())
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
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
				noobUser := helpers.RandomUsername()
				noobPassword := helpers.RandomPassword()
				session := helpers.CF("create-user", noobUser, noobPassword)
				Eventually(session).Should(Exit(0))
				session = helpers.CF("auth", noobUser, noobPassword)
				Eventually(session).Should(Exit(0))
				newUser = helpers.RandomUsername()
				newPassword = helpers.RandomPassword()
			})

			It("fails with insufficient scope error", func() {
				session := helpers.CF("create-user", newUser, newPassword)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("Error creating user %s.", newUser))
				Expect(session.Err).To(Say("Insufficient scope for this resource"))
				Expect(session.Out).To(Say("FAILED"))
			})
		})

		Context("when the logged in user is authorized to create new users", func() {
			BeforeEach(func() {
				helpers.LoginCF()
			})

			Context("when passed invalid username", func() {
				DescribeTable("when passed funkyUsername",
					func(funkyUsername string) {
						session := helpers.CF("create-user", funkyUsername, helpers.RandomPassword())
						Eventually(session).Should(Exit(1))
						Expect(session.Out).To(Say("Error creating user %s.", funkyUsername))
						Expect(session.Err).To(Say("Username must match pattern: \\[\\\\p\\{L\\}\\+0\\-9\\+\\\\\\-_\\.@'!\\]\\+"))
						Expect(session.Out).To(Say("FAILED"))
					},

					Entry("fails when passed an emoji", "😀"),
					Entry("fails when passed a backtick", "`"),
				)

				Context("when the username is empty", func() {
					It("fails with a username must be provided error", func() {
						session := helpers.CF("create-user", "", helpers.RandomPassword())
						Eventually(session).Should(Exit(1))
						Expect(session.Out).To(Say("Error creating user ."))
						Expect(session.Err).To(Say("A username must be provided."))
						Expect(session.Out).To(Say("FAILED"))
					})
				})
			})

			Context("when the user passes in an origin flag", func() {
				Context("when the origin is UAA", func() {
					Context("when password is not present", func() {
						It("errors and prints usage", func() {
							newUser := helpers.RandomUsername()
							session := helpers.CF("create-user", newUser, "--origin", "UAA")
							Eventually(session).Should(Exit(1))
							Expect(session.Err).To(Say("Incorrect Usage: the required argument `PASSWORD` was not provided"))
							Expect(session.Out).To(Say("FAILED"))
							Expect(session.Out).To(Say("USAGE"))
						})
					})
				})
				Context("when the origin is the empty string", func() {
					Context("when password is not present", func() {
						It("errors and prints usage", func() {
							newUser := helpers.RandomUsername()
							session := helpers.CF("create-user", newUser, "--origin", "")
							Eventually(session).Should(Exit(1))
							Expect(session.Err).To(Say("Incorrect Usage: the required argument `PASSWORD` was not provided"))
							Expect(session.Out).To(Say("FAILED"))
							Expect(session.Out).To(Say("USAGE"))
						})
					})
				})

				Context("when the origin is not UAA or empty", func() {
					It("creates the new user in the specified origin", func() {
						newUser := helpers.RandomUsername()
						session := helpers.CF("create-user", newUser, "--origin", "ldap")
						Eventually(session).Should(Exit(0))
						Expect(session.Out).To(Say("Creating user %s...", newUser))
						Expect(session.Out).To(Say("OK"))
						Expect(session.Out).To(Say("TIP: Assign roles with 'cf set-org-role' and 'cf set-space-role'"))
					})
				})

				Context("when argument for flag is not present", func() {
					It("fails with incorrect usage error", func() {
						session := helpers.CF("create-user", helpers.RandomUsername(), "--origin")
						Eventually(session).Should(Exit(1))
						Expect(session.Err).To(Say("Incorrect Usage: expected argument for flag `--origin'"))
					})
				})
			})

			Context("when password is not present", func() {
				It("fails with incorrect usage error", func() {
					session := helpers.CF("create-user", helpers.RandomUsername())
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("Incorrect Usage: the required argument `PASSWORD` was not provided"))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Out).To(Say("USAGE"))
				})
			})

			Context("when the user already exists", func() {
				var (
					newUser     string
					newPassword string
				)

				BeforeEach(func() {
					newUser = helpers.RandomUsername()
					newPassword = helpers.RandomPassword()
					session := helpers.CF("create-user", newUser, newPassword)
					Eventually(session).Should(Exit(0))
				})

				It("fails with the user already exists message", func() {
					session := helpers.CF("create-user", newUser, newPassword)
					Eventually(session).Should(Exit(0))
					Expect(session.Err).To(Say("user %s already exists", newUser))
					Expect(session.Out).To(Say("OK"))
				})
			})

			Context("when the user does not already exist", func() {
				It("creates the new user", func() {
					newUser := helpers.RandomUsername()
					newPassword := helpers.RandomPassword()
					session := helpers.CF("create-user", newUser, newPassword)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say("Creating user %s...", newUser))
					Expect(session.Out).To(Say("OK"))
					Expect(session.Out).To(Say("TIP: Assign roles with 'cf set-org-role' and 'cf set-space-role'"))
				})
			})
		})
	})
})
