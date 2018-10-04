package isolated

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

var _ = Describe("create-space", func() {
	var spaceName string

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
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE` was not provided"))
			expectHelpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not set up correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "create-space", "some-space")
		})
	})

	BeforeEach(func() {
		spaceName = helpers.NewSpaceName()
	})

	PWhen("logged in as a client", func() {
		var client, orgName string

		BeforeEach(func() {
			client = helpers.LoginCFWithClientCredentials()
			orgName = helpers.CreateAndTargetOrg()
		})

		It("successfully creates a space", func() {
			session := helpers.CF("create-space", spaceName)
			expectSuccessTextAndExitCode(session, client, orgName, spaceName)
		})
	})

	When("logged in as a user", func() {
		var user, orgName string

		BeforeEach(func() {
			user = helpers.LoginCF()
			orgName = helpers.CreateAndTargetOrg()
		})

		When("the space already exists", func() {
			BeforeEach(func() {
				session := helpers.CF("create-space", spaceName)
				Eventually(session).Should(Exit(0))
			})

			It("warns the user that the space already exists", func() {
				session := helpers.CF("create-space", spaceName)
				Eventually(session).Should(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, user))
				Eventually(session.Err).Should(Say(`Space %s already exists`, spaceName))
				Eventually(session).Should(Say(`OK\n`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the space does not exist yet", func() {
			When("a quota is not specified", func() {
				It("creates the space in the targeted org", func() {
					session := helpers.CF("create-space", spaceName)
					expectSuccessTextAndExitCode(session, user, orgName, spaceName)

					session = helpers.CF("space", spaceName)
					Eventually(session).Should(Say(`name:\s+%s`, spaceName))
					Eventually(session).Should(Say(`org:\s+%s`, orgName))
					Eventually(session).Should(Exit(0))
				})

				It("makes the user a space manager and space developer", func() {
					session := helpers.CF("create-space", spaceName)
					expectSuccessTextAndExitCode(session, user, orgName, spaceName)

					session = helpers.CF("space-users", orgName, spaceName)
					Eventually(session).Should(Say(`SPACE MANAGER\n\s+%s`, user))
					Eventually(session).Should(Say(`SPACE DEVELOPER\n\s+%s`, user))
					Eventually(session).Should(Exit(0))
				})
			})

			When("quota is specified", func() {
				var (
					quotaName string
					session   *Session
				)

				When("the quota exists", func() {
					BeforeEach(func() {
						quotaName = helpers.QuotaName()
						quotaSession := helpers.CF("create-space-quota", quotaName)
						Eventually(quotaSession).Should(Exit(0))
					})

					It("creates the space with the provided quota", func() {
						session = helpers.CF("create-space", spaceName, "-q", quotaName)
						expectSuccessTextAndExitCode(session, user, orgName, spaceName)
						session = helpers.CF("space", spaceName)
						Eventually(session).Should(Say(`name:\s+%s`, spaceName))
						Eventually(session).Should(Say(`space quota:\s+%s`, quotaName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the quota does not exist", func() {
					BeforeEach(func() {
						quotaName = "no-such-quota"
						session = helpers.CF("create-space", spaceName, "-q", quotaName)
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

				When("the org exists", func() {
					BeforeEach(func() {
						orgName = helpers.NewOrgName()
						orgSession := helpers.CF("create-org", orgName)
						Eventually(orgSession).Should(Exit(0))
					})

					It("creates the space in the specified org", func() {
						session = helpers.CF("create-space", spaceName, "-o", orgName)
						expectSuccessTextAndExitCode(session, user, orgName, spaceName)

						helpers.TargetOrg(orgName)
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

					It("fails with an error and does not create the space", func() {
						session = helpers.CF("create-space", spaceName, "-o", orgName)
						Eventually(session).Should(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, user))
						Eventually(session.Err).Should(Say(`Organization '%s' not found`, orgName))
						Eventually(session).Should(Say(`FAILED\n`))
						Eventually(session).Should(Exit(1))

						Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
					})
				})
			})

			When("the user is not authorized to create a space", func() {
				var user string

				BeforeEach(func() {
					user = helpers.SwitchToOrgRole(orgName, "OrgAuditor")
				})

				AfterEach(func() {
					helpers.ClearTarget()
					Expect(user).To(MatchRegexp(`^INTEGRATION-USER-[\da-f-]+$`))
					helpers.DeleteUser(user)
				})

				It("fails with an error telling the user that they are not authorized", func() {
					session := helpers.CF("create-space", spaceName, "-o", orgName)
					Eventually(session).Should(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, orgName, user))
					Eventually(session.Err).Should(Say(`You are not authorized to perform the requested action`))
					Eventually(session).Should(Say(`FAILED\n`))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
