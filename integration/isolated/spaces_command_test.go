package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("spaces command", func() {
	var orgName string

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("spaces", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("\\s+spaces - List all spaces in an org"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("\\s+cf spaces"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("\\s+target"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "spaces")
		})
	})

	Context("when the environment is setup correctly", func() {
		var username string

		BeforeEach(func() {
			username = helpers.LoginCF()
			helpers.CreateOrg(orgName)
			helpers.TargetOrg(orgName)
		})

		Context("when there are no spaces", func() {
			It("displays no spaces found", func() {
				session := helpers.CF("spaces")
				Eventually(session).Should(Say("Getting spaces in org %s as %s\\.\\.\\.", orgName, username))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("No spaces found\\."))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when there are multiple spaces", func() {
			var spaceName1, spaceName2, spaceName3, spaceName4, spaceName5, spaceName6 string

			BeforeEach(func() {
				spaceName1 = helpers.PrefixedRandomName("INTEGRATION-SPACE-DEF")
				spaceName2 = helpers.PrefixedRandomName("INTEGRATION-SPACE-XYZ")
				spaceName3 = helpers.PrefixedRandomName("INTEGRATION-SPACE-jop")
				spaceName4 = helpers.PrefixedRandomName("INTEGRATION-SPACE-ABC")
				spaceName5 = helpers.PrefixedRandomName("INTEGRATION-SPACE-123")
				spaceName6 = helpers.PrefixedRandomName("INTEGRATION-SPACE--")
				helpers.CreateSpace(spaceName1)
				helpers.CreateSpace(spaceName2)
				helpers.CreateSpace(spaceName3)
				helpers.CreateSpace(spaceName4)
				helpers.CreateSpace(spaceName5)
				helpers.CreateSpace(spaceName6)
			})

			It("displays a list of all spaces in the org in alphabetical order", func() {
				session := helpers.CF("spaces")
				Eventually(session).Should(Say("Getting spaces in org %s as %s\\.\\.\\.", orgName, username))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("name"))
				Eventually(session).Should(Say("%s\n%s\n%s\n%s\n%s\n%s", spaceName6, spaceName5, spaceName4, spaceName1, spaceName3, spaceName2))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
