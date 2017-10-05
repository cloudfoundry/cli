package isolated

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
			session := setup()
			//Eventually(session).Should(Say("push - Envoyer")) // TODO: Uncomment when language files have been updated
			//Eventually(session).Should(Say("SYNTAXE :"))
			//Eventually(session).Should(Say("Envoyez par commande push"))
			//Eventually(session).Should(Say("-i\\s+Nombre d'instances"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when the locale is set in the config", func() *Session {
			session := helpers.CF("config", "--locale", "fr-FR")
			Eventually(session).Should(Exit(0))

			return helpers.CF("push", "--help")
		}),

		Entry("when the the config and LANG environment variable is set, it uses config", func() *Session {
			session := helpers.CF("config", "--locale", "fr-FR")
			Eventually(session).Should(Exit(0))

			return helpers.CFWithEnv(map[string]string{"LANG": "es-ES"}, "push", "--help")
		}),

		Entry("when the the LANG environment variable is set", func() *Session {
			return helpers.CFWithEnv(map[string]string{"LANG": "fr-FR"}, "push", "--help")
		}),

		Entry("when the the LC_ALL environment variable is set", func() *Session {
			return helpers.CFWithEnv(map[string]string{"LC_ALL": "fr-FR"}, "push", "--help")
		}),

		Entry("when the the LC_ALL and LANG environment variables are set, it uses LC_ALL", func() *Session {
			return helpers.CFWithEnv(map[string]string{"LC_ALL": "fr-FR", "LANG": "es-ES"}, "push", "--help")
		}),

		Entry("when the the config, LC_ALL, and LANG is set, it uses config", func() *Session {
			session := helpers.CF("config", "--locale", "fr-FR")
			Eventually(session).Should(Exit(0))

			return helpers.CFWithEnv(map[string]string{"LC_ALL": "ja-JP", "LANG": "es-ES"}, "push", "--help")
		}),
	)

	DescribeTable("defaults to English",
		func(setup func() *Session) {
			session := setup()
			Eventually(session).Should(Say("push - Push a new app or sync changes to an existing app"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when the the LANG and LC_ALL environment variable is not set", func() *Session {
			return helpers.CF("push", "--help")
		}),

		Entry("when the the LANG environment variable is set to a non-supported langauge", func() *Session {
			return helpers.CFWithEnv(map[string]string{"LANG": "jj-FF"}, "push", "--help")
		}),

		Entry("when the the LC_ALL environment variable is set to a non-supported langauge", func() *Session {
			return helpers.CFWithEnv(map[string]string{"LC_ALL": "jj-FF"}, "push", "--help")
		}),
	)
})
