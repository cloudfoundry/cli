package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("login command", func() {
	Describe("Target Organization", func() {
		When("only a single organization is available", func() {
			var orgName string

			BeforeEach(func() {
				helpers.LoginCF()
				helpers.DeleteAllOrgs()
				orgName = helpers.NewOrgName()
				helpers.CreateOrg(orgName)
			})

			It("logs the user in and targets the organization automatically", func() {
				apiURL := helpers.GetAPI()
				username, password := helpers.GetCredentials()
				session := helpers.CF("login", "-u", username, "-p", password, "-a", apiURL)
				Eventually(session).Should(Exit(0))

				targetSession := helpers.CF("target")
				Eventually(targetSession).Should(Exit(0))
				Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
			})
		})
	})
})
