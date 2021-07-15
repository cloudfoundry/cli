package isolated

import (
	"fmt"

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
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		helpers.CreateOrgAndSpace(orgName, spaceName)
	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(orgName)
	})

	Describe("help text and argument validation", func() {
		When("--help flag is unset", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unset-space-role", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say("NAME:"))
				Expect(session).To(Say("unset-space-role - Remove a space role from a user"))
				Expect(session).To(Say("USAGE:"))
				Expect(session).To(Say("cf unset-space-role USERNAME ORG SPACE ROLE"))
				Expect(session).To(Say(`cf unset-space-role USERNAME ORG SPACE ROLE \[--client\]`))
				Expect(session).To(Say(`cf unset-space-role USERNAME ORG SPACE ROLE \[--origin ORIGIN\]`))
				Expect(session).To(Say("ROLES:"))
				Expect(session).To(Say("SpaceManager - Invite and manage users, and enable features for a given space"))
				Expect(session).To(Say("SpaceDeveloper - Create and manage apps and services, and see logs and reports"))
				Expect(session).To(Say("SpaceAuditor - View logs, reports, and settings on this space"))
				Expect(session).To(Say(`SpaceSupporter \[Beta role, subject to change\] - Manage app lifecycle and service bindings`))
				Expect(session).To(Say("OPTIONS:"))
				Expect(session).To(Say(`--client\s+Remove space role from a client-id of a \(non-user\) service account`))
				Expect(session).To(Say(`--origin\s+Indicates the identity provider to be used for authentication`))
				Expect(session).To(Say("SEE ALSO:"))
				Expect(session).To(Say("set-space-role, space-users"))
			})
		})

		When("the role type does not exist", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("unset-space-role", "some-user", "some-org", "some-space", "NotARealRole")
				Eventually(session.Err).Should(Say(`Incorrect Usage: ROLE must be "SpaceManager", "SpaceDeveloper", "SpaceAuditor" or "SpaceSupporter"`))
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
				session := helpers.CF("curl", "-X", "POST", "v3/users", "-d", fmt.Sprintf(`{"guid":"%s"}`, clientID))
				Eventually(session).Should(Exit(0))
			})

			When("the client exists and is affiliated with the active user's org", func() {
				BeforeEach(func() {
					session := helpers.CF("set-space-role", clientID, orgName, spaceName, "SpaceAuditor", "--client")
					Eventually(session).Should(Exit(0))
					privilegedUsername = helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceManager")
				})

				It("unsets the space role for the client", func() {
					session := helpers.CF("unset-space-role", clientID, orgName, spaceName, "SpaceAuditor", "--client")
					Eventually(session).Should(Say("Removing role SpaceAuditor from user %s in org %s / space %s as %s...", clientID, orgName, spaceName, privilegedUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

			})

			When("the active user lacks permissions to look up clients", func() {
				BeforeEach(func() {
					helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceManager")
				})

				It("prints an appropriate error and exits 1", func() {
					session := helpers.CF("unset-space-role", "cf_smoke_tests", orgName, spaceName, "SpaceAuditor", "--client")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("User '%s' does not exist.", "cf_smoke_tests"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the targeted client does not exist", func() {
				var badClientID string

				BeforeEach(func() {
					badClientID = helpers.NewUsername()
				})

				It("prints an appropriate error and exits 1", func() {
					session := helpers.CF("unset-space-role", badClientID, orgName, spaceName, "SpaceAuditor")
					Eventually(session).Should(Say("Removing role SpaceAuditor from user %s in org %s / space %s as %s...", badClientID, orgName, spaceName, privilegedUsername))
					Eventually(session.Err).Should(Say("User '%s' does not exist.", badClientID))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the user exists", func() {
			var username string

			BeforeEach(func() {
				username, _ = helpers.CreateUser()
				session := helpers.CF("set-space-role", username, orgName, spaceName, "spaceauditor")
				Eventually(session).Should(Exit(0))
			})

			When("the passed role type is lowercase", func() {
				It("unsets the space role for the user", func() {
					session := helpers.CF("unset-space-role", "-v", username, orgName, spaceName, "spaceauditor")
					Eventually(session).Should(Say("Removing role SpaceAuditor from user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			It("unsets the space role for the user", func() {
				session := helpers.CF("unset-space-role", username, orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("Removing role SpaceAuditor from user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("the user does not have the role to delete", func() {
				It("is idempotent", func() {
					session := helpers.CF("unset-space-role", username, orgName, spaceName, "SpaceDeveloper")
					Eventually(session).Should(Say("Removing role SpaceDeveloper from user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
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
					session := helpers.CF("unset-space-role", username, orgName, spaceName, "SpaceAuditor")
					Eventually(session).Should(Say("Removing role SpaceAuditor from user %s in org %s / space %s as %s...", username, orgName, spaceName, privilegedUsername))
					Eventually(session.Err).Should(Say("Ambiguous user. User with username '%s' exists in the following origins: cli-oidc-provider, uaa. Specify an origin to disambiguate.", username))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the user does not exist", func() {
			It("prints an appropriate error and exits 1", func() {
				session := helpers.CF("unset-space-role", "not-exists", orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("Removing role SpaceAuditor from user not-exists in org %s / space %s as %s...", orgName, spaceName, privilegedUsername))
				Eventually(session.Err).Should(Say("User 'not-exists' does not exist."))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the logged in user does not have permission to write to the space", func() {
		var username string

		BeforeEach(func() {
			username, _ = helpers.CreateUser()
			session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceAuditor")
			Eventually(session).Should(Exit(0))
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
			Eventually(session.Err).Should(Say("User '%s' does not exist.", username))
			Eventually(session).Should(Exit(1))
		})
	})
})
