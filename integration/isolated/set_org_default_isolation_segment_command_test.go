package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-org-default-isolation-segment command", func() {
	var orgName string
	var isoSegName string

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		isoSegName = helpers.IsolationSegmentName()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
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

	Context("when the environment is not set-up correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("set-org-default-isolation-segment", orgName, isoSegName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("set-org-default-isolation-segment", orgName, isoSegName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set-up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.LoginCF()
			userName, _ = helpers.GetCredentials()
		})

		Context("when the org does not exist", func() {
			It("fails with org not found message", func() {
				session := helpers.CF("set-org-default-isolation-segment", orgName, isoSegName)
				Eventually(session).Should(Say("Setting isolation segment %s to default on org %s as %s\\.\\.\\.", isoSegName, orgName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Organization '%s' not found\\.", orgName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the space exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
			})

			Context("when the isolation segment does not exist", func() {
				It("fails with isolation segment not found message", func() {
					session := helpers.CF("set-org-default-isolation-segment", orgName, isoSegName)
					Eventually(session).Should(Say("Setting isolation segment %s to default on org %s as %s\\.\\.\\.", isoSegName, orgName, userName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Isolation segment '%s' not found\\.", isoSegName))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the isolation segment exists", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-isolation-segment", isoSegName)).Should(Exit(0))
				})

				Context("when the isolation segment is entitled to the organization", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("enable-org-isolation", orgName, isoSegName)).Should(Exit(0))
					})

					It("displays OK", func() {
						session := helpers.CF("set-org-default-isolation-segment", orgName, isoSegName)
						Eventually(session).Should(Say("Setting isolation segment %s to default on org %s as %s\\.\\.\\.", isoSegName, orgName, userName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("In order to move running applications to this isolation segment, they must be restarted\\."))
						Eventually(session).Should(Exit(0))
					})

					Context("when the isolation segment is already set as the org's default", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("set-org-default-isolation-segment", orgName, isoSegName)).Should(Exit(0))
						})

						It("displays OK", func() {
							session := helpers.CF("set-org-default-isolation-segment", orgName, isoSegName)
							Eventually(session).Should(Say("Setting isolation segment %s to default on org %s as %s\\.\\.\\.", isoSegName, orgName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("In order to move running applications to this isolation segment, they must be restarted\\."))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})
})
