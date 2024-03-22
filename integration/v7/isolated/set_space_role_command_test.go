package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-space-role command", func() {
	Describe("help text and argument validation", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("set-space-role", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say("NAME:"))
				Expect(session).To(Say("set-space-role - Assign a space role to a user"))
				Expect(session).To(Say("USAGE:"))
				Expect(session).To(Say("cf set-space-role USERNAME ORG SPACE ROLE"))
				Expect(session).To(Say(`cf set-space-role USERNAME ORG SPACE ROLE \[--client\]`))
				Expect(session).To(Say(`cf set-space-role USERNAME ORG SPACE ROLE \[--origin ORIGIN\]`))
				Expect(session).To(Say("ROLES:"))
				Expect(session).To(Say("SpaceManager - Invite and manage users, and enable features for a given space"))
				Expect(session).To(Say("SpaceDeveloper - Create and manage apps and services, and see logs and reports"))
				Expect(session).To(Say("SpaceAuditor - View logs, reports, and settings on this space"))
				Expect(session).To(Say(`SpaceSupporter \[Beta role, subject to change\] - Manage app lifecycle and service bindings`))
				Expect(session).To(Say("OPTIONS:"))
				Expect(session).To(Say(`--client\s+Assign a space role to a client-id of a \(non-user\) service account`))
				Expect(session).To(Say(`--origin\s+Indicates the identity provider to be used for authentication`))
				Expect(session).To(Say("SEE ALSO:"))
				Expect(session).To(Say("space-users, unset-space-role"))
			})
		})

		When("the role type is invalid", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("set-space-role", "some-user", "some-org", "some-space", "NotARealRole")
				Eventually(session.Err).Should(Say(`Incorrect Usage: ROLE must be "SpaceManager", "SpaceDeveloper", "SpaceAuditor" or "SpaceSupporter"`))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too few arguments are passed", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("set-space-role", "not-enough", "arguments")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SPACE` and `ROLE` were not provided"))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too many arguments are passed", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("set-space-role", "some-user", "some-org", "some-space", "SpaceAuditor", "some-extra-argument")
				Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "some-extra-argument"`))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Describe("command behavior", func() {
		var (
			privilegedUsername string
			orgName            string
			spaceName          string
		)

		BeforeEach(func() {
			privilegedUsername = helpers.LoginCF()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("logged in as a privileged user", func() {
			When("the --client flag is passed", func() {
				var clientID string

				BeforeEach(func() {
					clientID, _ = helpers.SkipIfClientCredentialsNotSet()
				})

				When("the client exists", func() {
					It("sets the org role for the client", func() {
						session := helpers.CF("set-space-role", clientID, orgName, spaceName, "SpaceAuditor", "--client")
						Eventually(session).Should(Say("Assigning role SpaceAuditor to user %s in org %s / space %s as %s...", clientID, orgName, spaceName, privilegedUsername))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})

					When("the client is not authorized to look up clients in UAA", func() {
						BeforeEach(func() {
							helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceManager")
						})

						It("prints an appropriate error and exits 1", func() {
							session := helpers.CF("set-space-role", clientID, orgName, spaceName, "SpaceAuditor", "--client", "-v")
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("You are not authorized to perform the requested action."))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				When("the targeted client does not exist", func() {
					var badClientID string

					BeforeEach(func() {
						badClientID = "nonexistent-client"
					})

					It("fails with an appropriate error message", func() {
						session := helpers.CF("set-space-role", badClientID, orgName, spaceName, "SpaceAuditor", "--client")
						Eventually(session.Err).Should(Say("Users cannot be assigned roles in a space if they do not have a role in that space's organization."))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the user exists", func() {
				var username string

				BeforeEach(func() {
					username, _ = helpers.CreateUser()
				})

				When("the passed role is lowercase", func() {
					It("sets the space role for the user", func() {
						session := helpers.CF("set-space-role", username, orgName, spaceName, "spaceauditor")
						Eventually(session).Should(Say("Assigning role SpaceAuditor to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				It("sets the space role for the user", func() {
					session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceAuditor")
					Eventually(session).Should(Say("Assigning role SpaceAuditor to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

				When("the user already has the desired role", func() {
					BeforeEach(func() {
						session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceDeveloper")
						Eventually(session).Should(Say("Assigning role SpaceDeveloper to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
						Eventually(session).Should(Exit(0))
					})

					It("is idempotent", func() {
						session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceDeveloper")
						Eventually(session).Should(Say("Assigning role SpaceDeveloper to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the org does not exist", func() {
					It("prints an appropriate error and exits 1", func() {
						session := helpers.CF("set-space-role", username, "invalid-org", spaceName, "SpaceAuditor")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Organization 'invalid-org' not found."))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the space does not exist", func() {
					It("prints an appropriate error and exits 1", func() {
						session := helpers.CF("set-space-role", username, orgName, "invalid-space", "SpaceAuditor")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Space 'invalid-space' not found."))
						Eventually(session).Should(Exit(1))
					})
				})

				When("there are multiple users with the same username but different origins", func() {
					BeforeEach(func() {
						session := helpers.CF("create-user", username, "--origin", helpers.NonUAAOrigin)
						Eventually(session).Should(Exit(0))
					})

					AfterEach(func() {
						session := helpers.CF("delete-user", username, "--origin", helpers.NonUAAOrigin, "-f")
						Eventually(session).Should(Exit(0))
					})

					It("returns an error and asks the user to use the --origin flag", func() {
						session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceManager")
						Eventually(session).Should(Say("Assigning role SpaceManager to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
						Eventually(session.Err).Should(Say("Ambiguous user. User with username '%s' exists in the following origins: cli-oidc-provider, uaa. Specify an origin to disambiguate.", username))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the user does not exist", func() {
				It("prints an appropriate error and exits 1", func() {
					session := helpers.CF("set-space-role", "not-exists", orgName, spaceName, "SpaceAuditor")
					Eventually(session).Should(Say("Assigning role SpaceAuditor to user not-exists in org %s / space %s as %s...", orgName, spaceName, privilegedUsername))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("No user exists with the username 'not-exists'."))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the logged in user does not have permission to write to the space", func() {
			var username string

			BeforeEach(func() {
				username, _ = helpers.CreateUser()
				helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceAuditor")
			})

			It("prints out the error message from CC API and exits 1", func() {
				session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the logged in user has insufficient permissions to see the user", func() {
			var username string

			BeforeEach(func() {
				username, _ = helpers.CreateUser()
				helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceManager")
			})

			It("prints out the error message from CC API and exits 1", func() {
				session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceAuditor", "-v")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Users cannot be assigned roles in a space if they do not have a role in that space's organization."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the logged in user has insufficient permissions to create roles in the space", func() {
			var userInOrg string

			BeforeEach(func() {
				userInOrg, _ = helpers.CreateUser()
				Eventually(helpers.CF("set-org-role", userInOrg, orgName, "OrgAuditor")).Should(Exit(0))
				helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceAuditor")
			})

			It("prints out the error message from CC API and exits 1", func() {
				session := helpers.CF("set-space-role", userInOrg, orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
