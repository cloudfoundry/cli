package isolated

import (
	"os/exec"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Suggest Command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	When("a command is provided that is almost a command", func() {
		It("gives suggestions", func() {
			command := exec.Command("cf", "logn")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("'logn' is not a registered command. See 'cf help -a'"))
			Eventually(session).Should(Say("Did you mean?"))
			Eventually(session).Should(Exit(1))

			Eventually(session.Out.Contents()).Should(ContainSubstring("login"))
			Eventually(session.Out.Contents()).Should(ContainSubstring("logs"))
		})
	})

	When("a command is provided that is not even close", func() {
		It("gives suggestions", func() {
			command := exec.Command("cf", "zzz")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("'zzz' is not a registered command. See 'cf help -a'"))
			Consistently(session).ShouldNot(Say("Did you mean?"))
			Eventually(session).Should(Exit(1))
		})
	})
})
