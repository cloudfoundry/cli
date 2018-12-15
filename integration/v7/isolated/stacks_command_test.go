package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("stacks command", func() {
	var (
		orgName   string
		spaceName string
	)
	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
	})
	When("environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})
		It("lists the stacks", func() {
			session := helpers.CF("stacks")

			Eventually(session).Should(Say(`name\s+description`))
			Eventually(session).Should(Say(`cflinuxfs\d+\s+Cloud Foundry Linux`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("stacks", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`stacks - List all stacks \(a stack is a pre-built file system, including an operating system, that can run apps\)`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf stacks`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`app, push`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("environment is not set up correctly", func() {
		It("displays an error and exits 1", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "stacks")
		})
	})
})
