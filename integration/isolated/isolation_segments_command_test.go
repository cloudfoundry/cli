package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("isolation-segments command", func() {
	var isolationSegmentName string

	BeforeEach(func() {
		isolationSegmentName = helpers.IsolationSegmentName()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
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

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("isolation-segments")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("isolation-segments")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		// TODO: Pending until revoke-isolation-segments is done (for cleanup)
		PContext("when there are no isolation segments", func() {
			It("returns an ok and displays just the shared isolation segment", func() {
				userName, _ := helpers.GetCredentials()
				session := helpers.CF("isolation-segments")
				Eventually(session).Should(Say("Getting isolation segments as %s...", userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("name\\s+orgs"))
				Eventually(session).Should(Say("shared"))
				Consistently(session).ShouldNot(Say("[a-zA-Z]+"))
			})
		})

		// TODO: Pending until revoke-isolation-segments is done (for cleanup)
		PContext("when there are some isolation segments", func() {
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

				isolationSegment1 = helpers.IsolationSegmentName()
				isolationSegment2 = helpers.IsolationSegmentName()
				isolationSegment3 = helpers.IsolationSegmentName()

				Eventually(helpers.CF("create-isolation-segment", isolationSegment1)).Should(Exit(0))
				Eventually(helpers.CF("create-isolation-segment", isolationSegment2)).Should(Exit(0))
				Eventually(helpers.CF("enable-org-isolation", isolationSegment2, org1)).Should(Exit(0))
				Eventually(helpers.CF("create-isolation-segment", isolationSegment3)).Should(Exit(0))
				Eventually(helpers.CF("enable-org-isolation", isolationSegment3, org1)).Should(Exit(0))
				Eventually(helpers.CF("enable-org-isolation", isolationSegment3, org2)).Should(Exit(0))
			})

			// TODO: Delete this and add it to cleanup script after #138303919
			AfterEach(func() {
				Eventually(helpers.CF("delete-isolation-segment", "-f", isolationSegment1)).Should(Exit(0))
				Eventually(helpers.CF("delete-isolation-segment", "-f", isolationSegment2)).Should(Exit(0))
				Eventually(helpers.CF("delete-isolation-segment", "-f", isolationSegment3)).Should(Exit(0))
			})

			It("returns an ok and displays the table", func() {
				userName, _ := helpers.GetCredentials()
				session := helpers.CF("isolation-segments")
				Eventually(session).Should(Say("Getting isolation segments as %s...", userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("name\\s+orgs"))
				Eventually(session).Should(Say("shared"))
				Eventually(session).Should(Say("%s\\s+", isolationSegment1))
				Eventually(session).Should(Say("%s\\s+%s", isolationSegment2, org1))
				Eventually(session).Should(Say("%s\\s+%s, %s", isolationSegment3, org1, org2))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
