package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("reset-space-isolation-segment command", func() {
	var organizationName string
	var spaceName string

	BeforeEach(func() {
		organizationName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("reset-space-isolation-segment", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("reset-space-isolation-segment - Reset the space's isolation segment to the org default"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf reset-space-isolation-segment SPACE_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org, restart, space"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		When("no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("reset-space-isolation-segment", spaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("reset-space-isolation-segment", spaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is no org set", func() {
			BeforeEach(func() {
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("reset-space-isolation-segment", spaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
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
				session := helpers.CF("reset-space-isolation-segment", spaceName)
				Eventually(session).Should(Say("Resetting isolation segment assignment of space %s in org %s as %s...", spaceName, organizationName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Space '%s' not found.", spaceName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space exists", func() {
			BeforeEach(func() {
				helpers.CreateSpace(spaceName)
				isolationSegmentName := helpers.NewIsolationSegmentName()
				Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
				Eventually(helpers.CF("enable-org-isolation", organizationName, isolationSegmentName)).Should(Exit(0))
				Eventually(helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)).Should(Exit(0))
			})

			When("there is no default org isolation segment", func() {
				It("resets the space isolation segment to the shared isolation segment", func() {
					session := helpers.CF("reset-space-isolation-segment", spaceName)
					Eventually(session).Should(Say("Resetting isolation segment assignment of space %s in org %s as %s...", spaceName, organizationName, userName))

					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Applications in this space will be placed in the platform default isolation segment."))
					Eventually(session).Should(Say("Running applications need a restart to be moved there."))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space", spaceName)
					Eventually(session).Should(Say(`(?m)isolation segment:\s*$`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("there is a default org isolation segment", func() {
				var orgIsolationSegmentName string

				BeforeEach(func() {
					orgIsolationSegmentName = helpers.NewIsolationSegmentName()
					Eventually(helpers.CF("create-isolation-segment", orgIsolationSegmentName)).Should(Exit(0))
					Eventually(helpers.CF("enable-org-isolation", organizationName, orgIsolationSegmentName)).Should(Exit(0))
					Eventually(helpers.CF("set-org-default-isolation-segment", organizationName, orgIsolationSegmentName)).Should(Exit(0))
				})

				It("resets the space isolation segment to the default org isolation segment", func() {
					session := helpers.CF("reset-space-isolation-segment", spaceName)
					Eventually(session).Should(Say("Resetting isolation segment assignment of space %s in org %s as %s...", spaceName, organizationName, userName))

					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Applications in this space will be placed in isolation segment %s.", orgIsolationSegmentName))
					Eventually(session).Should(Say("Running applications need a restart to be moved there."))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space", spaceName)
					Eventually(session).Should(Say(`isolation segment:\s+%s`, orgIsolationSegmentName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
