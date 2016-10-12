package integration

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Version", func() {
	Context("when the -v option is provided without additional fields", func() {
		It("displays the version", func() {
			command := exec.Command("cf", "-v")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(Say("cf version"))
			Eventually(session).Should(Exit(0))
		})
	})
})
