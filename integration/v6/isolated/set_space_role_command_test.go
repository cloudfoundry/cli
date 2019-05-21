package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-space-role command", func() {
	Describe("help text and argument validation", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("set-space-role", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("set-space-role - Assign a space role to a user"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf set-space-role USERNAME ORG SPACE ROLE"))
				Eventually(session).Should(Say("ROLES:"))
				Eventually(session).Should(Say("'SpaceManager' - Invite and manage users, and enable features for a given space"))
				Eventually(session).Should(Say("'SpaceDeveloper' - Create and manage apps and services, and see logs and reports"))
				Eventually(session).Should(Say("'SpaceAuditor' - View logs, reports, and settings on this space"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--client\s+Treat USERNAME as the client-id of a \(non-user\) service account`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space-users"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the role does not exist", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("set-space-role", "some-user", "some-org", "some-space", "NotARealRole")
				Eventually(session.Err).Should(Say(`Incorrect Usage: ROLE must be "SpaceManager", "SpaceDeveloper" and "SpaceAuditor"`))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`\s+set-space-role - Assign a space role to a user`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf set-space-role USERNAME ORG SPACE ROLE`))
				Eventually(session).Should(Say(`ROLES:`))
				Eventually(session).Should(Say(`\s+'SpaceManager' - Invite and manage users, and enable features for a given space`))
				Eventually(session).Should(Say(`\s+'SpaceDeveloper' - Create and manage apps and services, and see logs and reports`))
				Eventually(session).Should(Say(`\s+'SpaceAuditor' - View logs, reports, and settings on this space`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--client\s+Treat USERNAME as the client-id of a \(non-user\) service account`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`\s+space-users`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too few arguments are passed", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("set-space-role", "not-enough", "arguments")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SPACE` and `ROLE` were not provided"))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`\s+set-space-role - Assign a space role to a user`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf set-space-role USERNAME ORG SPACE ROLE`))
				Eventually(session).Should(Say(`ROLES:`))
				Eventually(session).Should(Say(`\s+'SpaceManager' - Invite and manage users, and enable features for a given space`))
				Eventually(session).Should(Say(`\s+'SpaceDeveloper' - Create and manage apps and services, and see logs and reports`))
				Eventually(session).Should(Say(`\s+'SpaceAuditor' - View logs, reports, and settings on this space`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--client\s+Treat USERNAME as the client-id of a \(non-user\) service account`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`\s+space-users`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too many arguments are passed", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("set-space-role", "some-user", "some-org", "some-space", "SpaceAuditor", "some-extra-argument")
				Eventually(session).Should(Say("Incorrect Usage. Requires USERNAME, ORG, SPACE, ROLE as arguments"))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`\s+set-space-role - Assign a space role to a user`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf set-space-role USERNAME ORG SPACE ROLE`))
				Eventually(session).Should(Say(`ROLES:`))
				Eventually(session).Should(Say(`\s+'SpaceManager' - Invite and manage users, and enable features for a given space`))
				Eventually(session).Should(Say(`\s+'SpaceDeveloper' - Create and manage apps and services, and see logs and reports`))
				Eventually(session).Should(Say(`\s+'SpaceAuditor' - View logs, reports, and settings on this space`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--client\s+Treat USERNAME as the client-id of a \(non-user\) service account`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("logged in as admin", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			orgName = ReadOnlyOrg
			spaceName = ReadOnlySpace
		})

		When("the --client flag is passed", func() {
			var clientID string

			BeforeEach(func() {
				clientID, _ = helpers.SkipIfClientCredentialsNotSet()
			})

			When("the client exists", func() {
				It("sets the org role for the client", func() {
					helpers.SkipIfClientCredentialsTestMode()
					session := helpers.CF("set-space-role", clientID, orgName, spaceName, "SpaceAuditor", "--client")
					Eventually(session).Should(Say("Assigning role RoleSpaceAuditor to user %s in org %s / space %s as foo...", clientID, orgName, spaceName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

				When("the active user lacks permissions to look up clients", func() {
					BeforeEach(func() {
						helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceManager")
					})

					It("prints an appropriate error and exits 1", func() {
						session := helpers.CF("set-space-role", clientID, orgName, spaceName, "SpaceAuditor", "--client")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Say("Server error, status code: 403: Access is denied.  You do not have privileges to execute this command."))
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
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("Client nonexistent-client not found"))
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
					helpers.SkipIfClientCredentialsTestMode()
					session := helpers.CF("set-space-role", username, orgName, spaceName, "spaceauditor")
					Eventually(session).Should(Say("Assigning role RoleSpaceAuditor to user %s in org %s / space %s as foo...", username, orgName, spaceName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			It("sets the space role for the user", func() {
				helpers.SkipIfClientCredentialsTestMode()
				session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("Assigning role RoleSpaceAuditor to user %s in org %s / space %s as foo...", username, orgName, spaceName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("the logged in user has insufficient permissions", func() {
				BeforeEach(func() {
					helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceAuditor")
				})

				It("prints out the error message from CC API and exits 1", func() {
					session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceAuditor")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("Server error, status code: 403, error code: 10003, message: You are not authorized to perform the requested action"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the user already has the desired role", func() {
				BeforeEach(func() {
					helpers.SkipIfClientCredentialsTestMode()
					session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceDeveloper")
					Eventually(session).Should(Say("Assigning role RoleSpaceDeveloper to user %s in org %s / space %s as foo...", username, orgName, spaceName))
					Eventually(session).Should(Exit(0))
				})

				It("is idempotent", func() {
					session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceDeveloper")
					Eventually(session).Should(Say("Assigning role RoleSpaceDeveloper to user %s in org %s / space %s as foo...", username, orgName, spaceName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the org does not exist", func() {
				It("prints an appropriate error and exits 1", func() {
					session := helpers.CF("set-space-role", username, "invalid-org", spaceName, "SpaceAuditor")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("Organization invalid-org not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the space does not exist", func() {
				It("prints an appropriate error and exits 1", func() {
					session := helpers.CF("set-space-role", username, orgName, "invalid-space", "SpaceAuditor")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("Space invalid-space not found"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the user does not exist", func() {
			It("prints an appropriate error and exits 1", func() {
				helpers.SkipIfClientCredentialsTestMode()
				session := helpers.CF("set-space-role", "not-exists", orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("Assigning role RoleSpaceAuditor to user not-exists in org %s / space %s as foo...", orgName, spaceName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Server error, status code: 404, error code: 20003, message: The user could not be found: not-exists"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
