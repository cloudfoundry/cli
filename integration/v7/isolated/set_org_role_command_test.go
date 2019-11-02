package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-org-role command", func() {
	Context("Help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("set-org-role", "USER ADMIN", "Assign an org role to a user"))
			})

			It("displays the help information", func() {
				session := helpers.CF("set-org-role", "-h")
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`\s+set-org-role - Assign an org role to a user`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf set-org-role USERNAME ORG ROLE`))
				Eventually(session).Should(Say(`\s+cf set-org-role USERNAME ORG ROLE \[--client CLIENT\]`))
				Eventually(session).Should(Say(`\s+cf set-org-role USERNAME ORG ROLE \[--origin ORIGIN\]`))
				Eventually(session).Should(Say(`ROLES:`))
				Eventually(session).Should(Say(`\s+OrgManager - Invite and manage users, select and change plans, and set spending limits`))
				Eventually(session).Should(Say(`\s+BillingManager - Create and manage the billing account and payment info`))
				Eventually(session).Should(Say(`\s+OrgAuditor - Read-only access to org info and reports`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--client\s+Assign an org role to a client-id of a \(non-user\) service account`))
				Eventually(session).Should(Say(`--origin\s+Indicates the identity provider to be used for authentication`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`\s+org-users, set-space-role`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("reports that the user is not logged in", func() {
			session := helpers.CF("set-org-role", "some-user", "some-org", "BillingManager")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the user is logged in", func() {
		var orgName string
		var username string
		var currentUsername string

		BeforeEach(func() {
			currentUsername = helpers.LoginCF()
			orgName = ReadOnlyOrg
			username, _ = helpers.CreateUser()
		})

		When("the --client flag is passed", func() {
			When("the targeted user is actually a client", func() {
				var clientID string

				BeforeEach(func() {
					clientID, _ = helpers.SkipIfClientCredentialsNotSet()
				})

				It("sets the org role for the client", func() {
					session := helpers.CF("set-org-role", clientID, orgName, "OrgManager", "--client")
					Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as %s...", clientID, orgName, currentUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

				When("the active user lacks permissions to look up clients", func() {
					BeforeEach(func() {
						helpers.SwitchToOrgRole(orgName, "OrgManager")
					})

					It("prints an appropriate error and exits 1", func() {
						session := helpers.CF("set-org-role", "notaclient", orgName, "OrgManager", "--client")
						Eventually(session.Err).Should(Say("Invalid user. Ensure that the user exists and you have access to it."))
						Eventually(session).Should(Say("FAILED"))
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
					session := helpers.CF("set-org-role", badClientID, orgName, "OrgManager", "--client")
					Eventually(session.Err).Should(Say("Invalid user. Ensure that the user exists and you have access to it."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the --origin flag is passed", func() {
			When("the targeted user does not exist in the given origin", func() {
				var (
					targetUser string
					origin     string
				)

				BeforeEach(func() {
					targetUser = "some-user"
					origin = "some-origin"
				})

				It("fails with an appropriate error message", func() {
					session := helpers.CF("set-org-role", targetUser, orgName, "OrgManager", "--origin", origin)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("No user exists with the username 'some-user' and origin 'some-origin'."))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the org and user both exist", func() {
			When("the passed role is all lowercase", func() {
				It("sets the org role for the user", func() {
					session := helpers.CF("set-org-role", username, orgName, "orgauditor")
					Eventually(session).Should(Say("Assigning role OrgAuditor to user %s in org %s as %s...", username, orgName, currentUsername))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			It("sets the org role for the user", func() {
				session := helpers.CF("set-org-role", username, orgName, "OrgAuditor")
				Eventually(session).Should(Say("Assigning role OrgAuditor to user %s in org %s as %s...", username, orgName, currentUsername))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("the logged in user has insufficient permissions", func() {
				BeforeEach(func() {
					helpers.SwitchToOrgRole(orgName, "OrgAuditor")
				})

				It("prints out the error message from CC API and exits 1", func() {
					session := helpers.CF("set-org-role", username, orgName, "OrgAuditor")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the user already has the desired role", func() {
				BeforeEach(func() {
					session := helpers.CF("set-org-role", username, orgName, "OrgManager")
					Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as %s...", username, orgName, currentUsername))
					Eventually(session).Should(Exit(0))
				})

				It("is idempotent", func() {
					session := helpers.CF("set-org-role", username, orgName, "OrgManager")
					Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as %s...", username, orgName, currentUsername))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the specified role is invalid", func() {
				It("prints a useful error, prints help text, and exits 1", func() {
					session := helpers.CF("set-org-role", username, orgName, "NotARealRole")
					Eventually(session.Err).Should(Say(`Incorrect Usage: ROLE must be "OrgManager", "BillingManager" and "OrgAuditor"`))
					Eventually(session).Should(Say(`NAME:`))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the org does not exist", func() {
			It("prints an appropriate error and exits 1", func() {
				session := helpers.CF("set-org-role", username, "not-exists", "OrgAuditor")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Organization 'not-exists' not found."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the user does not exist", func() {
			It("prints an appropriate error and exits 1", func() {
				session := helpers.CF("set-org-role", "not-exists", orgName, "OrgAuditor")
				Eventually(session).Should(Say("Assigning role OrgAuditor to user not-exists in org %s as %s...", orgName, currentUsername))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No user exists with the username 'not-exists' and origin 'uaa'."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
