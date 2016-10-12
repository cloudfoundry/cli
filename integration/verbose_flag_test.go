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
	DescribeTable("displays verbose output",
		func(command func() *exec.Cmd) {
			login := exec.Command("cf", "auth", "admin", "admin")
			loginSession, err := Start(login, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(loginSession).Should(Exit(0))

			cmd := command()
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("REQUEST:"))
			Eventually(session).Should(Say("GET /v2/organizations"))
			Eventually(session).Should(Say("RESPONSE:"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when the -v option is provided with additional command", func() *exec.Cmd {
			return exec.Command("cf", "-v", "orgs")
		}),

		Entry("when the CF_TRACE env variable is set", func() *exec.Cmd {
			cmd := exec.Command("cf", "orgs")
			cmd.Env = helpers.AddOrReplaceEnvironment("CF_TRACE", "true")
			return cmd
		}),
	)
})
