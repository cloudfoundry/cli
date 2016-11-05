package integration

import (
	helpers "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("internationalization", func() {
	DescribeTable("outputs help in different languages",
		func(setup func() *Session) {
			Skip("Pending discussion on internationalization support")
			session := setup()
			Eventually(session).Should(Say("push - Envoyer"))
			Eventually(session).Should(Say("SYNTAXE :"))
			Eventually(session).Should(Say("Envoyez par commande push"))
			Eventually(session).Should(Say("-i\\s+Nombre d'instances"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when the locale is set in the config", func() *Session {
			session := helpers.CF("config", "--locale", "fr-FR")
			Eventually(session).Should(Exit(0))

			return helpers.CF("push", "--help")
		}),

		Entry("when the the LANG environment variable is set", func() *Session {
			return helpers.CFWithEnv(map[string]string{"LANG": "fr-FR"}, "push", "--help")
		}),
	)
})
