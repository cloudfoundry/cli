package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("router-groups command", func() {
	Describe("help", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("router-groups", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("router-groups - List router groups"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf router-groups"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("create-domain, domains"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "router-groups")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			userName = helpers.LoginCF()
		})

		When("security groups exists", func() {
			It("displays the security groups exits 0", func() {
				session := helpers.CF("router-groups")

				Eventually(session).Should(Say(`Getting router groups as %s\.\.\.`, userName))
				Eventually(session).Should(Say(`name\s+type`))
				Eventually(session).Should(Say(`%s\s+%s`, "default-tcp", "tcp"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
