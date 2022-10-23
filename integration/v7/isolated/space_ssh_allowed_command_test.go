package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("space-ssh-allowed command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("space-ssh-allowed", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("space-ssh-allowed - Reports whether SSH is allowed in a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf space-ssh-allowed SPACE_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("allow-space-ssh, ssh-enabled, ssh"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("command behavior", func() {
		var (
			orgName   string
			spaceName string
			session   *Session
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		JustBeforeEach(func() {
			session = helpers.CF("space-ssh-allowed", spaceName)
		})

		It("reports that ssh is allowed", func() {
			Eventually(session).Should(Say(`ssh support is enabled in space '%s'`, spaceName))
			Eventually(session).Should(Exit(0))
		})

		When("ssh is not allowed", func() {
			BeforeEach(func() {
				session := helpers.CF("disallow-space-ssh", spaceName)
				Eventually(session).Should(Exit(0))
			})

			It("reports that ssh is not allowed", func() {
				Eventually(session).Should(Say(`ssh support is disabled in space '%s'`, spaceName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the space does not exist", func() {
			BeforeEach(func() {
				spaceName = "nonexistent-space"
			})

			It("displays a missing space error message and fails", func() {
				Eventually(session.Err).Should(Say("Space '%s' not found.", spaceName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
