package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = PDescribe("orgs command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("orgs", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("\\s+orgs - List all orgs"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("\\s+cf orgs"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("\\s+o"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when an API endpoint is not set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("displays an error and exits 1", func() {
				session := helpers.CF("orgs")
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when an API endpoint is set", func() {
			Context("when the user is not logged in", func() {
				It("displays an error and exits 1", func() {
					session := helpers.CF("orgs")
					Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		var username string

		BeforeEach(func() {
			username = helpers.LoginCF()
		})

		// This test is skipped because we have remaining manual orgs that we do
		// not want to remove. We can unskip this test when we create a dev
		// environment for our manual testing in this story: #151187003
		PContext("when there are no orgs", func() {
			It("displays no orgs found", func() {
				session := helpers.CF("orgs")
				Eventually(session).Should(Say("Getting orgs as %s\\.\\.\\.", username))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("No orgs found\\."))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when there are multiple orgs", func() {
			var orgName1 string
			var orgName2 string

			BeforeEach(func() {
				orgName1 = helpers.NewOrgName()
				orgName2 = helpers.NewOrgName()
				helpers.CreateOrg(orgName1)
				helpers.CreateOrg(orgName2)
			})

			It("displays a list of all orgs", func() {
				session := helpers.CF("orgs")
				Eventually(session).Should(Say("Getting orgs as %s\\.\\.\\.", username))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("name"))
				Eventually(session).Should(Say("%s", orgName1))
				Eventually(session).Should(Say("%s", orgName2))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
