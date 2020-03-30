package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("isolation-segments command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("isolation-segments", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("isolation-segments - List all isolation segments"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf isolation-segments"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-isolation-segment, enable-org-isolation"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "isolation-segments")
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("there are some isolation segments", func() {
			var isolationSegment1 string // No orgs assigned
			var isolationSegment2 string // One org assigned
			var isolationSegment3 string // Many orgs assigned
			var org1 string
			var org2 string

			BeforeEach(func() {
				org1 = helpers.NewOrgName()
				org2 = helpers.NewOrgName()
				helpers.CreateOrg(org1)
				helpers.CreateOrg(org2)

				isolationSegment1 = helpers.NewIsolationSegmentName()
				isolationSegment2 = helpers.NewIsolationSegmentName()
				isolationSegment3 = helpers.NewIsolationSegmentName()

				Eventually(helpers.CF("create-isolation-segment", isolationSegment1)).Should(Exit(0))
				Eventually(helpers.CF("create-isolation-segment", isolationSegment2)).Should(Exit(0))
				Eventually(helpers.CF("create-isolation-segment", isolationSegment3)).Should(Exit(0))
				Eventually(helpers.CF("enable-org-isolation", org1, isolationSegment2)).Should(Exit(0))
				Eventually(helpers.CF("enable-org-isolation", org1, isolationSegment3)).Should(Exit(0))
				Eventually(helpers.CF("enable-org-isolation", org2, isolationSegment3)).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(org1)
				helpers.QuickDeleteOrg(org2)
			})

			It("returns an ok and displays the table", func() {
				userName, _ := helpers.GetCredentials()
				session := helpers.CF("isolation-segments")
				Eventually(session).Should(Say("Getting isolation segments as %s...", userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`name\s+orgs`))
				Eventually(session).Should(Say("shared"))
				Eventually(session).Should(Say(`%s\s+`, isolationSegment1))
				Eventually(session).Should(Say(`%s\s+%s`, isolationSegment2, org1))
				Eventually(session).Should(Say(`%s\s+%s, %s`, isolationSegment3, org1, org2))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
