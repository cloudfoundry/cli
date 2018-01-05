package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("disable-org-isolation command", func() {
	var organizationName string
	var isolationSegmentName string

	BeforeEach(func() {
		organizationName = helpers.NewOrgName()
		isolationSegmentName = helpers.NewIsolationSegmentName()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("disable-org-isolation", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("disable-org-isolation - Revoke an organization's entitlement to an isolation segment"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf disable-org-isolation ORG_NAME SEGMENT_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("enable-org-isolation, isolation-segments"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "disable-org-isolation", "org-name", "isolation-segment-name")
		})

		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("disable-org-isolation", organizationName, isolationSegmentName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.11\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithV3Version("3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("disable-org-isolation", organizationName, isolationSegmentName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.11\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.LoginCF()
			userName, _ = helpers.GetCredentials()
		})

		Context("when the org does not exist", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
			})

			It("outputs an error and exits 1", func() {
				session := helpers.CF("disable-org-isolation", organizationName, isolationSegmentName)
				Eventually(session).Should(Say("Removing entitlement to isolation segment %s from org %s as %s...", isolationSegmentName, organizationName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Organization '%s' not found.", organizationName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the isolation segment does not exist", func() {
			It("outputs an error and exits 1", func() {
				session := helpers.CF("disable-org-isolation", organizationName, isolationSegmentName)
				Eventually(session).Should(Say("Removing entitlement to isolation segment %s from org %s as %s...", isolationSegmentName, organizationName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Isolation segment '%s' not found.", isolationSegmentName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the binding does not exist", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
				Eventually(helpers.CF("create-org", organizationName)).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(organizationName)
			})

			It("outputs a warning and exists 0", func() {
				session := helpers.CF("disable-org-isolation", organizationName, isolationSegmentName)
				Eventually(session).Should(Say("Removing entitlement to isolation segment %s from org %s as %s...", isolationSegmentName, organizationName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				// Tests idempotence
				session = helpers.CF("disable-org-isolation", organizationName, isolationSegmentName)
				Eventually(session).Should(Say("Removing entitlement to isolation segment %s from org %s as %s...", isolationSegmentName, organizationName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when everything exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
				Eventually(helpers.CF("create-org", organizationName)).Should(Exit(0))
				Eventually(helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(organizationName)
			})

			It("displays OK", func() {
				session := helpers.CF("disable-org-isolation", organizationName, isolationSegmentName)
				Eventually(session).Should(Say("Removing entitlement to isolation segment %s from org %s as %s...", isolationSegmentName, organizationName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
