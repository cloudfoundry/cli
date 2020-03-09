package global

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("security-groups command", func() {
	When("there are no security groups", func() {
		var (
			unprivilegedUsername string
			password             string
			runningGroups        []string
			stagingGroups        []string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			session := helpers.CF("running-security-groups")
			Eventually(session).Should(Exit(0))
			runningGroups = strings.Split(string(session.Out.Contents()), "\n")[2:]
			for _, runningGroup := range runningGroups {
				helpers.CF(`unbind-running-security-group`, runningGroup)
			}
			session = helpers.CF("staging-security-groups")
			Eventually(session).Should(Exit(0))
			stagingGroups = strings.Split(string(session.Out.Contents()), "\n")[2:]
			for _, stagingGroup := range stagingGroups {
				helpers.CF(`unbind-staging-security-group`, stagingGroup)
			}
			unprivilegedUsername, password = helpers.CreateUser()
			helpers.LogoutCF()
			helpers.LoginAs(unprivilegedUsername, password)
		})

		AfterEach(func() {
			helpers.LoginCF()
			for _, runningGroup := range runningGroups {
				helpers.CF(`bind-running-security-group`, runningGroup)
			}
			for _, stagingGroup := range stagingGroups {
				helpers.CF(`bind-staging-security-group`, stagingGroup)
			}
			helpers.DeleteUser(unprivilegedUsername)
		})

		It("displays no security groups found and exits 0", func() {
			session := helpers.CF("security-groups")
			Eventually(session).Should(Say(`Getting security groups as %s\.\.\.`, unprivilegedUsername))
			Eventually(session.Out).Should(Say("No security groups found."))
			Eventually(session).Should(Exit(0))
		})
	})
})
