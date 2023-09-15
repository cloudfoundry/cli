package isolated

import (
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("stacks command", func() {
	When("--help flag is set", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("stacks", "APPS", "List all stacks (a stack is a pre-built file system, including an operating system, that can run apps)"))
		})

		It("Displays command usage to output", func() {
			session := helpers.CF("stacks", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(regexp.QuoteMeta("stacks - List all stacks (a stack is a pre-built file system, including an operating system, that can run apps)")))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf stacks [--labels SELECTOR]")))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say("cf stacks"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf stacks --labels 'environment in (production,staging),tier in (backend)'")))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf stacks --labels 'env=dev,!chargeback-code,tier in (backend,worker)'")))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`--labels\s+Selector to filter stacks by labels`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("create-buildpack, delete-buildpack, rename-buildpack, stack, update-buildpack"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("environment is not set up correctly", func() {
		It("displays an error and exits 1", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "stacks")
		})
	})

	When("environment is set up correctly", func() {
		var stackName string
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(ReadOnlyOrg, ReadOnlySpace)
			stackName = helpers.NewStackName()
			helpers.CreateStack(stackName)
			helperSession := helpers.CF("set-label", "stack", stackName, "cool=ranch")

			Eventually(helperSession).Should(Exit(0))
			userName, _ = helpers.GetCredentials()
		})

		When("--labels flag is set", func() {

			It("lists the filtered stacks by the flag", func() {
				session := helpers.CF("stacks", "--labels", "cool in (ranch)")

				Eventually(session).Should(Exit(0))

				Expect(session).Should(Say(`Getting stacks as %s\.\.\.`, userName))
				Expect(session).Should(Say(`name\s+description`))
				Expect(session).Should(Say(`%s\s+CF CLI integration test stack, please delete`, stackName))

				Expect(session).ShouldNot(Say(`cflinuxfs\d+\s+Cloud Foundry Linux`))

			})
		})

		It("lists the stacks", func() {
			session := helpers.CF("stacks")

			Eventually(session).Should(Exit(0))

			Expect(session).Should(Say(`Getting stacks as %s\.\.\.`, userName))
			Expect(session).Should(Say(`name\s+description`))
			Expect(session).Should(Say(`cflinuxfs\d+\s+Cloud Foundry Linux`))
			Expect(session).Should(Say(`%s\s+CF CLI integration test stack, please delete`, stackName))
		})
	})
})
