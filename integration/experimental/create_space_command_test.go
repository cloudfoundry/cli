package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func expectHelpText(session *Session) {
	Eventually(session).Should(Say(`NAME:`))
	Eventually(session).Should(Say(`create-space - Create a space\n`))
	Eventually(session).Should(Say(`\n`))

	Eventually(session).Should(Say(`USAGE:`))
	Eventually(session).Should(Say(`cf create-space SPACE \[-o ORG\] \[-q SPACE_QUOTA\]\n`))
	Eventually(session).Should(Say(`\n`))

	Eventually(session).Should(Say(`OPTIONS:`))
	Eventually(session).Should(Say(`-o\s+Organization`))
	Eventually(session).Should(Say(`-q\s+Quota to assign to the newly created space`))
	Eventually(session).Should(Say(`\n`))

	Eventually(session).Should(Say(`SEE ALSO:`))
	Eventually(session).Should(Say(`set-space-isolation-segment, space-quotas, spaces, target`))
}

func expectSuccessTextAndExitCode(session *Session, user, orgName, spaceName string) {
	Eventually(session).Should(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, user))
	Eventually(session).Should(Say(`OK\n`))
	Eventually(session).Should(Say(`Assigning role SpaceManager to user %s in org %s / space %s as %s\.\.\.`, user, orgName, spaceName, user))
	Eventually(session).Should(Say(`OK\n`))
	Eventually(session).Should(Say(`Assigning role SpaceDeveloper to user %s in org %s / space %s as %s\.\.\.`, user, orgName, spaceName, user))
	Eventually(session).Should(Say(`OK\n\n`))
	Eventually(session).Should(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, orgName, spaceName))
	Eventually(session).Should(Exit(0))
}

func createSpaceWhenNotAuthorized(orgName, user string) {
	helpers.LogoutCF()
	spaceName := helpers.NewSpaceName()
	session := helpers.CF("create-space", spaceName)
	Eventually(session).Should(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, user))
	Eventually(session.Err).Should(Say(`You are not authorized to perform the requested action`))
	Eventually(session).Should(Say(`FAILED\n`))
	Eventually(session).Should(Exit(1))
}

var _ = PDescribe("create-space", func() {
	When("invoked with --help", func() {
		It("displays the help information", func() {
			session := helpers.CF("create-space", "--help")
			expectHelpText(session)
			Eventually(session).Should(Exit(0))
		})
	})

	When("invoked with no arguments", func() {
		It("shows an error and the help text", func() {
			session := helpers.CF("create-space")
			Eventually(session).Should(Say(`Incorrect Usage: the required argument 'SPACE' was not provided`))
			expectHelpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not set up correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, "", "create-space", "some-space")
		})
	})

	When("logged in as a client", func() {
		var client, orgName string

		BeforeEach(func() {
			client = helpers.LoginCFWithClientCredentials()
			orgName = helpers.CreateAndTargetOrg()
		})

		It("successfully creates a space", func() {
			spaceName := helpers.NewSpaceName()
			session := helpers.CF("create-space", spaceName)
			expectSuccessTextAndExitCode(session, client, orgName, spaceName)
		})
	})

	When("logged in as a user", func() {
		var user, orgName, spaceName string

		BeforeEach(func() {
			user = helpers.LoginCF()
			orgName = helpers.CreateAndTargetOrg()
		})

		When("the space already exists", func() {
			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				session := helpers.CF("create-space", spaceName)
				Eventually(session).Should(Exit(0))
			})

			It("warns the user that the space already exists", func() {
				session := helpers.CF("create-space", spaceName)
				Eventually(session).Should(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, user))
				Eventually(session).Should(Say(`OK\n`))
				Eventually(session.Err).Should(Say(`Space %s already exists`, spaceName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the space does not exist yet", func() {
			When("a quota is not specified", func() {
				var session *Session

				JustBeforeEach(func() {
					spaceName = helpers.NewSpaceName()
					session = helpers.CF("create-space", spaceName)
				})

				It("displays success output", func() {
					expectSuccessTextAndExitCode(session, user, orgName, spaceName)
				})

				It("creates the space in the targeted org", func() {
					session = helpers.CF("space", spaceName)
					Expect(session).To(Say(`name:\s+%s`, spaceName))
					Expect(session).To(Say(`org:\s+%s`, orgName))
					Eventually(session).Should(Exit(0))
				})

				It("makes the user a space manager and developer", func() {
					usersSession := helpers.CF("space-users", orgName, spaceName)
					Expect(usersSession).To(Say(`SPACE MANAGER\n\s+%s`, user))
					Expect(usersSession).To(Say(`SPACE DEVELOPER\n\s+%s`, user))
					Eventually(usersSession).Should(Exit(0))
				})
			})

			When("quota is specified", func() {
				var (
					quotaName string
					session   *Session
				)

				JustBeforeEach(func() {
					spaceName = helpers.NewSpaceName()
					session = helpers.CF("create-space", spaceName, "-q", quotaName)
				})

				When("the quota exists", func() {
					BeforeEach(func() {
						quotaName = helpers.QuotaName()
						quotaSession := helpers.CF("create-space-quota", quotaName)
						Eventually(quotaSession).Should(Exit(0))
					})

					It("displays success output", func() {
						expectSuccessTextAndExitCode(session, user, orgName, spaceName)
					})

					It("creates the space with the provided quota", func() {
						session = helpers.CF("space", spaceName)
						Expect(session).To(Say(`name:\s+%s`, spaceName))
						Expect(session).To(Say(`space quota:\s+%s`, quotaName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the quota does not exist", func() {
					BeforeEach(func() {
						quotaName = "no-such-quota"
					})

					It("fails with an error", func() {
						Eventually(session).Should(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, user))
						Eventually(session.Err).Should(Say(`Quota no-such-quota not found`))
						Eventually(session).Should(Say(`FAILED\n`))
						Eventually(session).Should(Exit(1))
					})

					It("does not create the space", func() {
						Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
					})
				})
			})

			When("org is specified", func() {
				var session *Session

				JustBeforeEach(func() {
					spaceName = helpers.NewSpaceName()
					session = helpers.CF("create-space", spaceName, "-o", orgName)
				})

				When("the org exists", func() {
					BeforeEach(func() {
						orgName = helpers.NewOrgName()
						orgSession := helpers.CF("create-org", orgName)
						Eventually(orgSession).Should(Exit(0))
					})

					It("displays success output", func() {
						expectSuccessTextAndExitCode(session, user, orgName, spaceName)
					})

					It("creates the space in the specified org", func() {
						session = helpers.CF("space", spaceName)
						Eventually(session).Should(Say(`name:\s+%s`, spaceName))
						Eventually(session).Should(Say(`org:\s+%s`, orgName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the org does not exist", func() {
					BeforeEach(func() {
						orgName = "no-such-org"
					})

					It("fails with an error", func() {
						Eventually(session).Should(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, user))
						Eventually(session.Err).Should(Say(`Org no-such-org does not exist or is not accessible`))
						Eventually(session).Should(Say(`FAILED\n`))
						Eventually(session).Should(Exit(1))
					})

					It("does not create the space", func() {
						Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
					})
				})
			})

			When("creating a space as a non-admin user", func() {
				AfterEach(func() {
					helpers.ClearTarget()
					helpers.DeleteUser(user)
				})

				It("does not allow a user with no roles to create a space", func() {
					user = helpers.SwitchToNoRole()
					createSpaceWhenNotAuthorized(orgName, user)
				})

				Describe("users of another space", func() {
					var existingSpace string

					BeforeEach(func() {
						existingSpace = helpers.NewSpaceName()
						session := helpers.CF("create-space", existingSpace)
						Eventually(session).Should(Exit(0))
					})

					It("does not allow a space developer to create a space", func() {
						user = helpers.SwitchToSpaceRole(orgName, existingSpace, "SpaceDeveloper")
						createSpaceWhenNotAuthorized(orgName, user)
					})

					It("does not allow a space auditor to create a space", func() {
						user = helpers.SwitchToSpaceRole(orgName, existingSpace, "SpaceAuditor")
						createSpaceWhenNotAuthorized(orgName, user)
					})

					It("does not allow a space manager to create a space", func() {
						user = helpers.SwitchToSpaceRole(orgName, existingSpace, "SpaceManager")
						createSpaceWhenNotAuthorized(orgName, user)
					})
				})

				Describe("users of this org", func() {
					It("does not allow an org auditor so create a space", func() {
						user = helpers.SwitchToOrgRole(orgName, "OrgAuditor")
						createSpaceWhenNotAuthorized(orgName, user)
					})

					It("does not allow an billing manager so create a space", func() {
						user = helpers.SwitchToOrgRole(orgName, "BillingManager")
						createSpaceWhenNotAuthorized(orgName, user)
					})

					It("allows an org manager to create a space", func() {
						user = helpers.SwitchToOrgRole(orgName, "OrgManager")
						spaceName = helpers.NewSpaceName()
						session := helpers.CF("create-space", spaceName)
						expectSuccessTextAndExitCode(session, user, orgName, spaceName)
					})
				})
			})
		})
	})
})
