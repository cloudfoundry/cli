package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("disallow-space-ssh command", func() {
	var (
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("disallow-space-ssh", "SPACES", "Disallow SSH access for the space"))
			})

			It("displays command usage to output", func() {
				session := helpers.CF("disallow-space-ssh", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("disallow-space-ssh - Disallow SSH access for the space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf disallow-space-ssh SPACE"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("disable-ssh, space-ssh-allowed, ssh, ssh-enabled"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})
		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the space name is not provided", func() {
			It("tells the user that the space name is required, prints help text, and exits 1", func() {
				session := helpers.CF("disallow-space-ssh")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space does not exist", func() {
			It("displays space not found and exits 1", func() {
				invalidSpaceName := "invalid-space-name"
				session := helpers.CF("disallow-space-ssh", invalidSpaceName)

				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Space %s not found", invalidSpaceName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("when ssh has not been disallowed yet", func() {
			It("disallows ssh for the space", func() {
				session := helpers.CF("disallow-space-ssh", spaceName)

				Eventually(session).Should(Say(`Disabling ssh support for space '%s'\.\.\.`, spaceName))

				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("space-ssh-allowed", spaceName)
				Eventually(session).Should(Say("ssh support is disabled in space '%s'", spaceName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("ssh was previously disallowed for the space", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("disallow-space-ssh", spaceName)).Should(Exit(0))
			})

			It("informs the user and exits 0", func() {
				session := helpers.CF("disallow-space-ssh", spaceName)

				Eventually(session).Should(Say("ssh support is already disabled in space '%s'", spaceName))

				session = helpers.CF("space-ssh-allowed", spaceName)
				Eventually(session).Should(Say("ssh support is disabled in space '%s'", spaceName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
