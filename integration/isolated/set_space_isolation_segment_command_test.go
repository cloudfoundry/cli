package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
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
		Context("when --help flag is set", func() {
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

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "set-space-isolation-segment", "space-name", "isolation-seg-name")
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
				session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
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
				session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
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
			helpers.CreateOrg(organizationName)
			helpers.TargetOrg(organizationName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(organizationName)
		})

		Context("when the space does not exist", func() {
			It("fails with space not found message", func() {
				session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
				Eventually(session).Should(Say("Updating isolation segment of space %s in org %s as %s\\.\\.\\.", spaceName, organizationName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Space '%s' not found.", spaceName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the space exists", func() {
			BeforeEach(func() {
				helpers.CreateSpace(spaceName)
			})

			Context("when the isolation segment does not exist", func() {
				It("fails with isolation segment not found message", func() {
					session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
					Eventually(session).Should(Say("Updating isolation segment of space %s in org %s as %s\\.\\.\\.", spaceName, organizationName, userName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Isolation segment '%s' not found.", isolationSegmentName))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the isolation segment exists", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
				})

				Context("when the isolation segment is entitled to the organization", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)).Should(Exit(0))
					})

					It("displays OK", func() {
						session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
						Eventually(session).Should(Say("Updating isolation segment of space %s in org %s as %s\\.\\.\\.", spaceName, organizationName, userName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("In order to move running applications to this isolation segment, they must be restarted."))
						Eventually(session).Should(Exit(0))
					})

					Context("when the isolation is already set to space", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)).Should(Exit(0))
						})

						It("displays OK", func() {
							session := helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)
							Eventually(session).Should(Say("Updating isolation segment of space %s in org %s as %s\\.\\.\\.", spaceName, organizationName, userName))
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
