package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-space-isolation-segment command", func() {
	var organizationName string
	var spaceName string
	var isolationSegmentName string

	BeforeEach(func() {
		organizationName = helpers.NewOrgName()
		isolationSegmentName = helpers.NewIsolationSegmentName()
		spaceName = helpers.NewSpaceName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("set-space-isolation-segment", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("set-space-isolation-segment - Assign the isolation segment for a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf set-space-isolation-segment SPACE_NAME SEGMENT_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org, reset-space-isolation-segment, restart, set-org-default-isolation-segment, space"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "set-space-isolation-segment", "space-name", "isolation-seg-name")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.LoginCF()
			userName, _ = helpers.GetCredentials()
			helpers.CreateOrg(organizationName)
			helpers.TargetOrg(organizationName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(organizationName)
		})

		When("the space does not exist", func() {
			It("fails with space not found message", func() {
				session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
				Eventually(session).Should(Say(`Updating isolation segment of space %s in org %s as %s\.\.\.`, spaceName, organizationName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Space '%s' not found.", spaceName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space exists", func() {
			BeforeEach(func() {
				helpers.CreateSpace(spaceName)
			})

			When("the isolation segment does not exist", func() {
				It("fails with isolation segment not found message", func() {
					session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
					Eventually(session).Should(Say(`Updating isolation segment of space %s in org %s as %s\.\.\.`, spaceName, organizationName, userName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Isolation segment '%s' not found.", isolationSegmentName))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the isolation segment exists", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
				})

				When("the isolation segment is entitled to the organization", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)).Should(Exit(0))
					})

					It("displays OK", func() {
						session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
						Eventually(session).Should(Say(`Updating isolation segment of space %s in org %s as %s\.\.\.`, spaceName, organizationName, userName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("In order to move running applications to this isolation segment, they must be restarted."))
						Eventually(session).Should(Exit(0))
					})

					When("the isolation is already set to space", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)).Should(Exit(0))
						})

						It("displays OK", func() {
							session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
							Eventually(session).Should(Say(`Updating isolation segment of space %s in org %s as %s\.\.\.`, spaceName, organizationName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("In order to move running applications to this isolation segment, they must be restarted."))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})
})
