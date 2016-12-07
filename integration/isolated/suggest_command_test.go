package isolated

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Suggest Command", func() {
	Context("when a command is provided that is almost a command", func() {
		It("gives suggestions", func() {
			command := exec.Command("cf", "logn")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session.Out).Should(Say("'logn' is not a registered command. See 'cf help'"))
			Eventually(session.Out).Should(Say("Did you mean?"))
			Eventually(session.Out.Contents()).Should(ContainSubstring("login"))
			Eventually(session.Out.Contents()).Should(ContainSubstring("logs"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when a command is provided that is not even close", func() {
		It("gives suggestions", func() {
			command := exec.Command("cf", "zzz")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session.Out).Should(Say("'zzz' is not a registered command. See 'cf help'"))
			Consistently(session.Out).ShouldNot(Say("Did you mean?"))
			Eventually(session).Should(Exit(1))
		})
	})
})
