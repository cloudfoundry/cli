package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-isolation-segment command", func() {
	var isolationSegmentName string

	BeforeEach(func() {
		isolationSegmentName = helpers.NewIsolationSegmentName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("create-isolation-segment", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-isolation-segment - Create an isolation segment"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf create-isolation-segment SEGMENT_NAME"))
				Eventually(session).Should(Say("NOTES:"))
				Eventually(session).Should(Say("The isolation segment name must match the placement tag applied to the Diego cell."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("enable-org-isolation, isolation-segments"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-isolation-segment", "isolation-seg-name")
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("the isolation segment does not exist", func() {
			It("creates the isolation segment", func() {
				session := helpers.CF("create-isolation-segment", isolationSegmentName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating isolation segment %s as %s...", isolationSegmentName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the isolation segment already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
			})

			It("returns an ok", func() {
				session := helpers.CF("create-isolation-segment", isolationSegmentName)
				Eventually(session.Err).Should(Say("Isolation segment '%s' already exists", isolationSegmentName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
