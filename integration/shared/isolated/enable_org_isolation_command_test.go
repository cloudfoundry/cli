package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("enable-org-isolation command", func() {
	var organizationName string
	var isolationSegmentName string

	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
		organizationName = helpers.NewOrgName()
		isolationSegmentName = helpers.NewIsolationSegmentName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("enable-org-isolation", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("enable-org-isolation - Entitle an organization to an isolation segment"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf enable-org-isolation ORG_NAME SEGMENT_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-isolation-segment, isolation-segments, set-org-default-isolation-segment, set-space-isolation-segment"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "enable-org-isolation", "org-name", "segment-name")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.LoginCF()
			userName, _ = helpers.GetCredentials()
		})

		When("the isolation segment does not exist", func() {
			It("fails with isolation segment not found message", func() {
				session := helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)
				Eventually(session).Should(Say("Enabling isolation segment %s for org %s as %s...", isolationSegmentName, organizationName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Isolation segment '%s' not found.", isolationSegmentName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the isolation segment exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
			})

			When("the organization does not exist", func() {
				It("fails with organization not found message", func() {
					session := helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)
					Eventually(session).Should(Say("Enabling isolation segment %s for org %s as %s...", isolationSegmentName, organizationName, userName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization '%s' not found.", organizationName))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the organization exists", func() {
				BeforeEach(func() {
					helpers.CreateOrg(organizationName)
					helpers.TargetOrg(organizationName)
				})

				AfterEach(func() {
					helpers.QuickDeleteOrg(organizationName)
				})

				It("displays OK", func() {
					session := helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)
					Eventually(session).Should(Say("Enabling isolation segment %s for org %s as %s...", isolationSegmentName, organizationName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

				When("the isolation is already enabled", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)).Should(Exit(0))
					})

					It("displays OK", func() {
						session := helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)
						Eventually(session).Should(Say("Enabling isolation segment %s for org %s as %s...", isolationSegmentName, organizationName, userName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
