package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("ssh-code command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("ssh-code", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("ssh-code - Get a one time password for ssh clients"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf ssh-code"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("curl, ssh"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "ssh-code")
		})
	})

	Context("when the environment is setup correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("returns a one time passcode for ssh", func() {
			session := helpers.CF("ssh-code")
			Eventually(session.Out).Should(Say("[A-Za-z0-9]+"))
			Eventually(session).Should(Exit(0))
		})
	})
})
