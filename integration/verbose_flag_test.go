package integration

import (
	"os/exec"

	helpers "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Verbose", func() {
	Context("v2 legacy", func() {
		DescribeTable("displays verbose output",
			func(command func() *Session) {
				login := exec.Command("cf", "auth", "admin", "admin")
				loginSession, err := Start(login, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(loginSession).Should(Exit(0))

				session := command()
				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("GET /v2/organizations"))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Exit(0))
			},

			Entry("when the -v option is provided with additional command", func() *Session {
				return helpers.CF("-v", "orgs")
			}),

			Entry("when the CF_TRACE env variable is set", func() *Session {
				return helpers.CFWithEnv(map[string]string{"CF_TRACE": "true"}, "orgs")
			}),
		)
	})

	Context("v2 refactor", func() {
		DescribeTable("displays verbose output",
			func(command func() *Session) {
				Skip("skip #133310639")
				orgName := helpers.PrefixedRandomName("testorg")
				spaceName := helpers.PrefixedRandomName("testspace")
				setupCF(orgName, spaceName)

				defer helpers.CF("delete-org", "-f", orgName)
				session := command()

				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("GET /v2/apps"))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Exit(1))
			},

			Entry("when the -v option is provided with additional command", func() *Session {
				return helpers.CF("-v", "unbind-service", "app-name", "service-instance")
			}),

			Entry("when the CF_TRACE env variable is set", func() *Session {
				return helpers.CFWithEnv(map[string]string{"CF_TRACE": "true"}, "unbind-service", "app-name", "service-instance")
			}),
		)
	})

	Context("v3", func() {
		DescribeTable("displays verbose output",
			func(command func() *Session) {
				orgName := helpers.PrefixedRandomName("testorg")
				spaceName := helpers.PrefixedRandomName("testspace")
				setupCF(orgName, spaceName)

				defer helpers.CF("delete-org", "-f", orgName)
				session := command()

				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("GET /v3/apps"))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Exit(1))
			},

			Entry("when the -v option is provided with additional command", func() *Session {
				return helpers.CF("-v", "run-task", "app", "echo")
			}),

			Entry("when the CF_TRACE env variable is set", func() *Session {
				return helpers.CFWithEnv(map[string]string{"CF_TRACE": "true"}, "run-task", "app", "echo")
			}),
		)
	})
})
