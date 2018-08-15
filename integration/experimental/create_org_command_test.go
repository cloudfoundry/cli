package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-org", func() {
	When("invoked with --help", func() {
		It("displays the help information", func() {
			session := helpers.CF("create-org", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`create-org - Create an org\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf create-org ORG\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`-q\s+Quota to assign to the newly created org \(excluding this option results in assignment of default quota\)`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`create-space, orgs, quotas, set-org-role`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not set up correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "create-org", "some-org")
		})
	})

	When("the environment is set up correctly", func() {
		var user string

		BeforeEach(func() {
			user = helpers.LoginCF()
		})

		When("the logged in user is not allowed to create orgs", func() {
			BeforeEach(func() {
				nonAdminUser := helpers.NewUsername()
				nonAdminPassword := helpers.NewPassword()

				session := helpers.CF("create-user", nonAdminUser, nonAdminPassword)
				Eventually(session).Should(Exit(0))

				env := map[string]string{
					"CF_USERNAME": nonAdminUser,
					"CF_PASSWORD": nonAdminPassword,
				}

				session = helpers.CFWithEnv(env, "auth")
				Eventually(session).Should(Exit(0))
			})

			XIt("fails with an insufficient scope error", func() {
				orgName := helpers.NewOrgName()
				session := helpers.CF("create-org", orgName)
				Eventually(session.Out).Should(Say("Error creating organization %s\\.", orgName))
				Eventually(session.Err).Should(Say("You are not authorized to perform the requested action\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the org does not exist yet", func() {
			It("creates the org", func() {
				orgName := helpers.NewOrgName()
				session := helpers.CF("create-org", orgName)

				Eventually(session.Out).Should(Say("Creating org %s as %s\\.\\.\\.", orgName, user))
				Eventually(session.Out).Should(Say("OK\\n\\n"))
				Eventually(session.Out).Should(Say("Assigning role OrgManager to user %s in org %s\\.\\.\\.", user, orgName))
				Eventually(session.Out).Should(Say("OK\\n\\n"))
				Eventually(session.Out).Should(Say(`TIP: Use 'cf target -o "%s"' to target new org`, orgName))
				Eventually(session).Should(Exit(0))
			})

			When("an existing quota is provided", func() {
				var quotaName string

				BeforeEach(func() {
					quotaName = helpers.QuotaName()
					session := helpers.CF("create-quota", quotaName)
					Eventually(session).Should(Exit(0))
				})

				XIt("creates the org with the provided quota", func() {
					orgName := helpers.NewOrgName()
					session := helpers.CF("create-org", orgName, "-q", quotaName)
					Eventually(session).Should(Exit(0))

					session = helpers.CF("org", orgName)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say("quota:\\s+%s", quotaName))
				})
			})

			When("a nonexistent quota is provided", func() {
				XIt("fails with an error and does not create the org", func() {
					orgName := helpers.NewOrgName()
					session := helpers.CF("create-org", orgName, "-q", "no-such-quota")
					Eventually(session.Err).Should(Say("FAILED\\n"))
					Eventually(session.Err).Should(Say("Quota no-such-quota not found"))
					Eventually(session).Should(Exit(1))

					Expect(helpers.CF("org", orgName)).To(Exit(1))
				})
			})
		})

		When("the org already exists", func() {
			var orgName string

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				session := helpers.CF("create-org", orgName)
				Eventually(session).Should(Exit(0))
			})

			XIt("warns the user that the org already exists", func() {
				session := helpers.CF("create-org", orgName)
				Eventually(session.Out).Should(Say("Creating org %s as %s\\.\\.\\.", orgName, user))
				Eventually(session.Out).Should(Say("OK\\n"))
				Eventually(session.Err).Should(Say("Org %s already exists", orgName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
