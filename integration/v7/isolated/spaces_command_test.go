package isolated

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
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
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("spaces", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("spaces - List all spaces in an org"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf spaces [--labels SELECTOR]")))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("cf spaces"))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf spaces --labels 'environment in (production,staging),tier in (backend)'")))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf spaces --labels 'env=dev,!chargeback-code,tier in (backend,worker)'")))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--labels\s+Selector to filter spaces by labels`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-space, set-space-role, space, space-users"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is setup correctly", func() {
		var username, spaceName1, spaceName2, spaceName3, spaceName4, spaceName5, spaceName6 string

		BeforeEach(func() {
			username = helpers.LoginCF()
			helpers.CreateOrg(orgName)
			helpers.TargetOrg(orgName)

			spaceName1 = helpers.PrefixedRandomName("INTEGRATION-SPACE-DEF")
			spaceName2 = helpers.PrefixedRandomName("INTEGRATION-SPACE-XYZ")
			spaceName3 = helpers.PrefixedRandomName("INTEGRATION-SPACE-jop")
			spaceName4 = helpers.PrefixedRandomName("INTEGRATION-SPACE-ABC")
			spaceName5 = helpers.PrefixedRandomName("INTEGRATION-SPACE-543")
			spaceName6 = helpers.PrefixedRandomName("INTEGRATION-SPACE-125")
			helpers.CreateSpace(spaceName1)
			helpers.CreateSpace(spaceName2)
			helpers.CreateSpace(spaceName3)
			helpers.CreateSpace(spaceName4)
			helpers.CreateSpace(spaceName5)
			helpers.CreateSpace(spaceName6)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the --labels flag is given", func() {

			BeforeEach(func() {
				Eventually(helpers.CF("set-label", "space", spaceName1, "environment=production", "tier=backend")).Should(Exit(0))
				Eventually(helpers.CF("set-label", "space", spaceName2, "environment=staging", "tier=frontend")).Should(Exit(0))
			})

			It("displays spaces with provided labels", func() {
				session := helpers.CF("spaces", "--labels", "environment in (production,staging),tier in (backend)")
				Eventually(session).Should(Say(`Getting spaces in org %s as %s\.\.\.`, orgName, username))
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(spaceName1))
				Expect(session).ShouldNot(Say(spaceName2))
				Expect(session).ShouldNot(Say(spaceName3))
				Expect(session).ShouldNot(Say(spaceName4))
				Expect(session).ShouldNot(Say(spaceName5))
				Expect(session).ShouldNot(Say(spaceName6))
			})

			It("displays no spaces when no labels match", func() {
				session := helpers.CF("spaces", "--labels", "environment in (production,staging),environment notin (production,staging)")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say("No spaces found"))
			})
		})
	})
})
