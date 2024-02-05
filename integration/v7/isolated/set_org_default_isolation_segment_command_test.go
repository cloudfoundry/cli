package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-org-default-isolation-segment command", func() {
	var (
		orgName              string
		isolationSegmentName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		isolationSegmentName = helpers.NewIsolationSegmentName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("set-org-default-isolation-segment", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("set-org-default-isolation-segment - Set the default isolation segment used for apps in spaces in an org"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf set-org-default-isolation-segment ORG_NAME SEGMENT_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org, set-space-isolation-segment"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "set-org-default-isolation-segment", "orgname", "segment-name")
		})
	})

	When("the environment is set-up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.LoginCF()
			userName, _ = helpers.GetCredentials()
		})

		When("the org does not exist", func() {
			It("fails with org not found message", func() {
				session := helpers.CF("set-org-default-isolation-segment", orgName, isolationSegmentName)
				Eventually(session).Should(Say(`Setting isolation segment %s to default on org %s as %s\.\.\.`, isolationSegmentName, orgName, userName))
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

			When("the isolation segment does not exist", func() {
				It("fails with isolation segment not found message", func() {
					session := helpers.CF("set-org-default-isolation-segment", orgName, isolationSegmentName)
					Eventually(session).Should(Say(`Setting isolation segment %s to default on org %s as %s\.\.\.`, isolationSegmentName, orgName, userName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Isolation segment '%s' not found\.`, isolationSegmentName))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the isolation segment exists", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
				})

				When("the isolation segment is entitled to the organization", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("enable-org-isolation", orgName, isolationSegmentName)).Should(Exit(0))
					})

					It("displays OK", func() {
						session := helpers.CF("set-org-default-isolation-segment", orgName, isolationSegmentName)
						Eventually(session).Should(Say(`Setting isolation segment %s to default on org %s as %s\.\.\.`, isolationSegmentName, orgName, userName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("TIP: Restart applications in this organization to relocate them to this isolation segment."))
						Eventually(session).Should(Exit(0))
					})

					When("the isolation segment is already set as the org's default", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("set-org-default-isolation-segment", orgName, isolationSegmentName)).Should(Exit(0))
						})

						It("displays OK", func() {
							session := helpers.CF("set-org-default-isolation-segment", orgName, isolationSegmentName)
							Eventually(session).Should(Say(`Setting isolation segment %s to default on org %s as %s\.\.\.`, isolationSegmentName, orgName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("TIP: Restart applications in this organization to relocate them to this isolation segment."))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})
})
