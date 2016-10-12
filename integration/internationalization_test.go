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

var _ = Describe("internationalization", func() {
	DescribeTable("outputs help in different languages",
		func(setup func() *exec.Cmd) {
			Skip("Pending discussion on internationalization support")
			cmd := setup()
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("push - Envoyer"))
			Eventually(session).Should(Say("SYNTAXE :"))
			Eventually(session).Should(Say("Envoyez par commande push"))
			Eventually(session).Should(Say("-i\\s+Nombre d'instances"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when the locale is set in the config", func() *exec.Cmd {
			cmd := exec.Command("cf", "config", "--locale", "fr-FR")
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(Exit(0))

			return exec.Command("cf", "push", "--help")
		}),

		Entry("when the the LANG environment variable is set", func() *exec.Cmd {
			cmd := exec.Command("cf", "push", "--help")
			cmd.Env = helpers.AddOrReplaceEnvironment("LANG", "fr-FR")
			return cmd
		}),
	)
})
