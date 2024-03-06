package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("reset-org-default-isolation-segment command", func() {
	var orgName string

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("reset-org-default-isolation-segment", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("reset-org-default-isolation-segment - Reset the default isolation segment used for apps in spaces of an org"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf reset-org-default-isolation-segment ORG_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org, restart"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "reset-org-default-isolation-segment", "org-name")
		})
	})

	When("the environment is set-up correctly", func() {
		var userName string
		var userOrgName string

		BeforeEach(func() {
			helpers.LoginCF()
			userName, _ = helpers.GetCredentials()
			userOrgName = helpers.NewOrgName()
			helpers.CreateOrg(userOrgName)
			helpers.TargetOrg(userOrgName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(userOrgName)
		})

		When("the org does not exist", func() {
			It("fails with org not found message", func() {
				session := helpers.CF("reset-org-default-isolation-segment", orgName)
				Eventually(session).Should(Say(`Resetting default isolation segment of org %s as %s\.\.\.`, orgName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Organization '%s' not found\.`, orgName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the org exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			When("the isolation segment is set as the org's default", func() {
				BeforeEach(func() {
					isolationSegmentName := helpers.NewIsolationSegmentName()
					Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
					Eventually(helpers.CF("enable-org-isolation", orgName, isolationSegmentName)).Should(Exit(0))
					Eventually(helpers.CF("set-org-default-isolation-segment", orgName, isolationSegmentName)).Should(Exit(0))
				})

				It("displays OK", func() {
					session := helpers.CF("reset-org-default-isolation-segment", orgName)
					Eventually(session).Should(Say(`Resetting default isolation segment of org %s as %s\.\.\.`, orgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Applications in spaces of this org that have no isolation segment assigned will be placed in the platform default isolation segment\.`))
					Eventually(session).Should(Say(`Running applications need a restart to be moved there\.`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the org has no default isolation segment", func() {
				It("displays OK", func() {
					session := helpers.CF("reset-org-default-isolation-segment", orgName)
					Eventually(session).Should(Say(`Resetting default isolation segment of org %s as %s\.\.\.`, orgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Applications in spaces of this org that have no isolation segment assigned will be placed in the platform default isolation segment\.`))
					Eventually(session).Should(Say(`Running applications need a restart to be moved there\.`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
