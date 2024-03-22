package isolated

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Suggest Command", func() {
	When("a command is provided that is almost a command", func() {
		It("gives suggestions", func() {
			command := exec.Command("cf", "logn")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)

			Eventually(session).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			Expect(session.Err).To(Say("'logn' is not a registered command. See 'cf help -a'"))
			Expect(session.Err).To(Say("Did you mean?"))

			Expect(session.Err.Contents()).To(ContainSubstring("login"))
			Expect(session.Err.Contents()).To(ContainSubstring("logs"))
		})
	})

	When("a command is provided that is not even close", func() {
		It("gives suggestions", func() {
			command := exec.Command("cf", "zzz")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)

			Eventually(session).Should(Exit(1))
			Expect(err).NotTo(HaveOccurred())

			Expect(session.Err).To(Say("'zzz' is not a registered command. See 'cf help -a'"))
			Expect(session.Err).ToNot(Say("Did you mean?"))
		})
	})
})
