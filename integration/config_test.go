package integration

import (
	"os/exec"

	helpers "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Config", func() {
	Describe("Enable Color", func() {
		Context("when color is enabled", func() {
			It("prints colors", func() {
				command := exec.Command("cf", "help")
				command.Env = helpers.AddOrReplaceEnvironment("CF_COLOR", "true")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(Say("\x1b\\[38;1m"))
			})
		})

		Context("when color is disabled", func() {
			It("does not print colors", func() {
				command := exec.Command("cf", "help")
				command.Env = helpers.AddOrReplaceEnvironment("CF_COLOR", "false")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(Exit(0))
				Expect(session).NotTo(Say("\x1b\\[38;1m"))
			})
		})
	})
})
