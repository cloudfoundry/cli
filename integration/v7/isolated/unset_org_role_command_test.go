package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = PDescribe("unset-org-role command", func() {
	var (
		privilegedUsername string
		orgName            string
	)

	BeforeEach(func() {
		privilegedUsername = helpers.LoginCF()
		orgName = ReadOnlyOrg
	})

	Describe("help text and argument validation", func() {
		When("--help flag is unset", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unset-org-role", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("unset-org-role - Remove an org role from a user"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf unset-org-role USERNAME ORG ROLE"))
				Eventually(session).Should(Say(`cf unset-org-role USERNAME ORG ROLE \[--client\]`))
				Eventually(session).Should(Say(`cf unset-org-role USERNAME ORG ROLE \[--origin ORIGIN\]`))
				Eventually(session).Should(Say("ROLES:"))
				Eventually(session).Should(Say("OrgManager - Invite and manage users, select and change plans, and set spending limits"))
				Eventually(session).Should(Say("BillingManager - Create and manage the billing account and payment info"))
				Eventually(session).Should(Say("OrgAuditor - Read-only access to org info and reports"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--client\s+Unassign an org role for a client-id of a \(non-user\) service account`))
				Eventually(session).Should(Say(`--origin\s+Indicates the identity provider to be used for authentication`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org-users, set-space-role"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the role does not exist", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("unset-org-role", "some-user", "some-org", "NotARealRole")
				Eventually(session.Err).Should(Say(`Incorrect Usage: ROLE must be "OrgManager", "BillingManager" and "OrgAuditor"`))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too few arguments are passed", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("unset-org-role", "not-enough-args")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `ORG` and `ROLE` were not provided"))
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("too many arguments are passed", func() {
			It("prints a useful error, prints help text, and exits 1", func() {
				session := helpers.CF("unset-org-role", "some-user", "some-org", "OrgAuditor", "some-extra-argument")
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

				session = helpers.CF("set-org-role", clientID, orgName, "OrgAuditor", "--client")
				Eventually(session).Should(Exit(0))
			})

			When("the client exists", func() {
				It("unsets the org role for the client", func() {
					session := helpers.CF("unset-org-role", clientID, orgName, "OrgAuditor", "--client")
					Eventually(session).Should(Say("Removing role OrgAuditor from user %s in org %s as %s...", clientID, orgName, privilegedUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the targeted client does not exist", func() {
				var badClientID string

				BeforeEach(func() {
					badClientID = "nonexistent-client"
				})

				It("fails with an appropriate error message", func() {
					session := helpers.CF("unset-org-role", badClientID, orgName, "OrgAuditor", "--client")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("User 'nonexistent-client' does not exist."))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the user exists", func() {
			var username string

			BeforeEach(func() {
				username, _ = helpers.CreateUser()
				session := helpers.CF("set-org-role", username, orgName, "orgauditor")
				Eventually(session).Should(Exit(0))
			})

			When("the passed role is lowercase", func() {
				It("unsets the org role for the user", func() {
					session := helpers.CF("unset-org-role", "-v", username, orgName, "orgauditor")
					Eventually(session).Should(Say("Removing role OrgAuditor from user %s in org %s as %s...", username, orgName, privilegedUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			It("unsets the org role for the user", func() {
				session := helpers.CF("unset-org-role", username, orgName, "OrgAuditor")
				Eventually(session).Should(Say("Removing role OrgAuditor from user %s in org %s as %s...", username, orgName, privilegedUsername))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("the user does not have the role to delete", func() {
				It("is idempotent", func() {
					session := helpers.CF("unset-org-role", username, orgName, "BillingManager")
					Eventually(session).Should(Say("Removing role BillingManager from user %s in org %s as %s...", username, orgName, privilegedUsername))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the org does not exist", func() {
				It("prints an appropriate error and exits 1", func() {
					session := helpers.CF("unset-org-role", username, "invalid-org", "OrgAuditor")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization 'invalid-org' not found."))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the user does not exist", func() {
			It("succeeds (idempotent case) and exits 0", func() {
				session := helpers.CF("unset-org-role", "not-exists", orgName, "OrgAuditor")
				Eventually(session).Should(Say("Removing role OrgAuditor from user not-exists in org %s as %s...", orgName, privilegedUsername))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the logged in user does not have permission to write to the org", func() {
		var username string

		BeforeEach(func() {
			username, _ = helpers.CreateUser()
			helpers.SwitchToOrgRole(orgName, "OrgAuditor")
		})

		It("prints out the error message from CC API and exits 1", func() {
			session := helpers.CF("unset-org-role", username, orgName, "OrgAuditor")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`User '%s' does not exist.`, username))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the logged in user has insufficient permissions to see the user", func() {
		var username string

		BeforeEach(func() {
			username, _ = helpers.CreateUser()
			helpers.SwitchToOrgRole(orgName, "OrgManager")
		})

		It("prints out the error message from CC API and exits 1", func() {
			session := helpers.CF("unset-org-role", username, orgName, "OrgAuditor", "-v")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("User '%s' does not exist.", username))
			Eventually(session).Should(Exit(1))
		})
	})
})
