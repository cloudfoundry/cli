package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("orgs command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("orgs", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+orgs - List all orgs`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf orgs`))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say(`\s+o`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`create-org, org, org-users`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "orgs")
		})
	})

	When("the environment is setup correctly", func() {
		var username string

		BeforeEach(func() {
			username = helpers.LoginCF()
		})

		When("there are multiple orgs", func() {
			var orgName1, orgName2, orgName3, orgName4, orgName5 string

			BeforeEach(func() {
				orgName1 = helpers.PrefixedRandomName("INTEGRATION-ORG-XYZ")
				orgName2 = helpers.PrefixedRandomName("INTEGRATION-ORG-456")
				orgName3 = helpers.PrefixedRandomName("INTEGRATION-ORG-ABC")
				orgName4 = helpers.PrefixedRandomName("INTEGRATION-ORG-123")
				orgName5 = helpers.PrefixedRandomName("INTEGRATION-ORG-ghi")
				helpers.CreateOrg(orgName1)
				helpers.CreateOrg(orgName2)
				helpers.CreateOrg(orgName3)
				helpers.CreateOrg(orgName4)
				helpers.CreateOrg(orgName5)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName1)
				helpers.QuickDeleteOrg(orgName2)
				helpers.QuickDeleteOrg(orgName3)
				helpers.QuickDeleteOrg(orgName4)
				helpers.QuickDeleteOrg(orgName5)
			})

			It("displays a list of all orgs", func() {
				session := helpers.CF("orgs")
				Eventually(session).Should(Say(`Getting orgs as %s\.\.\.`, username))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("name"))
				Eventually(session).Should(Say("%s", orgName4))
				Eventually(session).Should(Say("%s", orgName2))
				Eventually(session).Should(Say("%s", orgName3))
				Eventually(session).Should(Say("%s", orgName5))
				Eventually(session).Should(Say("%s", orgName1))
				Eventually(session).Should(Exit(0))
			})

			When("the --labels flag is given", func() {

				BeforeEach(func() {
					Eventually(helpers.CF("set-label", "org", orgName1, "environment=production", "tier=backend")).Should(Exit(0))
					Eventually(helpers.CF("set-label", "org", orgName2, "environment=staging", "tier=frontend")).Should(Exit(0))
				})

				It("displays only the organizations with labels that match the expression", func() {
					session := helpers.CF("orgs", "--labels", "environment in (production,staging),tier in (backend)")
					Eventually(session).Should(Exit(0))
					Expect(session).ShouldNot(Say(orgName2))
					Expect(session).Should(Say(orgName1))
				})

				When("the --labels selector is malformed", func() {
					It("errors", func() {
						session := helpers.CF("orgs", "--labels", "malformed in (")
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

	})
})
