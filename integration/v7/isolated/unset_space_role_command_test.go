package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unset-space-role command", func() {
	var (
		privilegedUsername string
		orgName            string
		spaceName          string
	)

	BeforeEach(func() {
		privilegedUsername = helpers.LoginCF()
		orgName = ReadOnlyOrg
		spaceName = ReadOnlySpace
	})

	Describe("help text and argument validation", func() {
		When("--help flag is unset", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unset-space-role", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("unset-space-role - Assign a space role to a user"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf unset-space-role USERNAME ORG SPACE ROLE"))
				Eventually(session).Should(Say(`cf unset-space-role USERNAME ORG SPACE ROLE \[--client\]`))
				Eventually(session).Should(Say(`cf unset-space-role USERNAME ORG SPACE ROLE \[--origin ORIGIN\]`))
				Eventually(session).Should(Say("ROLES:"))
				Eventually(session).Should(Say("SpaceManager - Invite and manage users, and enable features for a given space"))
				Eventually(session).Should(Say("SpaceDeveloper - Create and manage apps and services, and see logs and reports"))
				Eventually(session).Should(Say("SpaceAuditor - View logs, reports, and settings on this space"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--client\s+Assign a space role to a client-id of a \(non-user\) service account`))
				Eventually(session).Should(Say(`--origin\s+Indicates the identity provider to be used for authentication`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space-users"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the role does not exist", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("unset-space-role", "some-user", "some-org", "some-space", "NotARealRole")
				Eventually(session.Err).Should(Say(`Incorrect Usage: ROLE must be "SpaceManager", "SpaceDeveloper" and "SpaceAuditor"`))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too few arguments are passed", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("unset-space-role", "not-enough", "arguments")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SPACE` and `ROLE` were not provided"))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too many arguments are passed", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("unset-space-role", "some-user", "some-org", "some-space", "SpaceAuditor", "some-extra-argument")
				Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "some-extra-argument"`))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("logged in as a privileged user", func() {
		When("the --client flag is passed", func() {
			var clientID string

			BeforeEach(func() {
				clientID, _ = helpers.SkipIfClientCredentialsNotSet()
			})

			When("the client exists", func() {
				It("unsets the org role for the client", func() {
					session := helpers.CF("unset-space-role", clientID, orgName, spaceName, "SpaceAuditor", "--client")
					Eventually(session).Should(Say("Assigning role SpaceAuditor to user %s in org %s / space %s as %s...", clientID, orgName, spaceName, privilegedUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

				When("the active user lacks permissions to look up clients", func() {
					BeforeEach(func() {
						helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceManager")
					})

					It("prints an appropriate error and exits 1", func() {
						session := helpers.CF("unset-space-role", clientID, orgName, spaceName, "SpaceAuditor", "--client")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Users cannot be assigned roles in a space if they do not have a role in that space's organization."))
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
					session := helpers.CF("unset-space-role", badClientID, orgName, spaceName, "SpaceAuditor", "--client")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Invalid user. Ensure that the user exists and you have access to it."))
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
				It("unsets the space role for the user", func() {
					session := helpers.CF("unset-space-role", username, orgName, spaceName, "spaceauditor")
					Eventually(session).Should(Say("Assigning role SpaceAuditor to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			It("unsets the space role for the user", func() {
				session := helpers.CF("unset-space-role", username, orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("Assigning role SpaceAuditor to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("the user already has the desired role", func() {
				BeforeEach(func() {
					session := helpers.CF("unset-space-role", username, orgName, spaceName, "SpaceDeveloper")
					Eventually(session).Should(Say("Assigning role SpaceDeveloper to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
					Eventually(session).Should(Exit(0))
				})

				It("is idempotent", func() {
					session := helpers.CF("unset-space-role", username, orgName, spaceName, "SpaceDeveloper")
					Eventually(session).Should(Say("Assigning role SpaceDeveloper to user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the org does not exist", func() {
				It("prints an appropriate error and exits 1", func() {
					session := helpers.CF("unset-space-role", username, "invalid-org", spaceName, "SpaceAuditor")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization 'invalid-org' not found."))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the space does not exist", func() {
				It("prints an appropriate error and exits 1", func() {
					session := helpers.CF("unset-space-role", username, orgName, "invalid-space", "SpaceAuditor")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Space 'invalid-space' not found."))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the user does not exist", func() {
			It("prints an appropriate error and exits 1", func() {
				session := helpers.CF("unset-space-role", "not-exists", orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("Assigning role SpaceAuditor to user not-exists in org %s / space %s as %s...", orgName, spaceName, privilegedUsername))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No user exists with the username 'not-exists' and origin 'uaa'."))
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
			session := helpers.CF("unset-space-role", username, orgName, spaceName, "SpaceAuditor")
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
			session := helpers.CF("unset-space-role", username, orgName, spaceName, "SpaceAuditor", "-v")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Users cannot be assigned roles in a space if they do not have a role in that space's organization."))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the logged in user has insufficient permissions to create roles in the space", func() {
		var userInOrg string

		BeforeEach(func() {
			userInOrg, _ = helpers.CreateUser()
			Eventually(helpers.CF("unset-org-role", userInOrg, orgName, "OrgAuditor")).Should(Exit(0))
			helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceAuditor")
		})

		It("prints out the error message from CC API and exits 1", func() {
			session := helpers.CF("unset-space-role", userInOrg, orgName, spaceName, "SpaceAuditor")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
			Eventually(session).Should(Exit(1))
		})
	})
})
