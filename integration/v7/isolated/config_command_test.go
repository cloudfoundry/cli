package isolated

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("config command", func() {
	var helpText func(session *Session)

	BeforeEach(func() {
		helpText = func(session *Session) {
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`config - Write default values to the config`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf config \[--async-timeout TIMEOUT_IN_MINUTES\] \[--trace \(true | false | path/to/file\)\] \[--color \(true | false\)\] \[--locale \(LOCALE | CLEAR\)\]`))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`--async-timeout\s+Timeout in minutes for async HTTP requests`))
			Eventually(session).Should(Say(`--color\s+Enable or disable color in CLI output`))
			Eventually(session).Should(Say(`--locale\s+Set default locale. If LOCALE is 'CLEAR', previous locale is deleted.`))
			Eventually(session).Should(Say(`--trace\s+Trace HTTP requests by default. If a file path is provided then output will write to the file provided. If the file does not exist it will be created.`))
		}
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("shows the help text", func() {
				session := helpers.CF("config", "--help")
				helpText(session)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when no flags are given", func() {
		It("returns an error and displays the help text", func() {
			session := helpers.CF("config")
			Eventually(session.Err).Should(Say("Incorrect Usage: at least one flag must be provided"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	DescribeTable("allows setting async timeout to",
		func(timeout string) {
			session := helpers.CF("config", "--async-timeout", timeout)
			Eventually(session).Should(Say(`Setting values in config\.\.\.`))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		},

		Entry("integer", "10"),
	)

	When("the timeout provided is not an integer", func() {
		It("fails with the appropriate errors", func() {
			session := helpers.CF("config", "--async-timeout", "ten-seconds")
			Eventually(session.Err).Should(Say("Incorrect Usage: Timeout must be an integer greater than or equal to 1"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the timeout provided is 0 or less", func() {
		It("fails with the appropriate errors", func() {
			session := helpers.CF("config", "--async-timeout", "0")
			Eventually(session.Err).Should(Say("Incorrect Usage: Timeout must be an integer greater than or equal to 1"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	DescribeTable("allows setting color to",
		func(color string) {
			session := helpers.CF("config", "--color", color)
			Eventually(session).Should(Exit(0))
		},

		Entry("true", "true"),
		Entry("false", "false"),
	)

	When("the color argument provided is not a boolean", func() {
		It("fails with the appropriate errors", func() {
			session := helpers.CF("config", "--color", "Teal")
			Eventually(session.Err).Should(Say("Incorrect Usage: COLOR must be \"true\" or \"false\""))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	DescribeTable("allows setting locale to",
		func(locale string) {
			session := helpers.CF("config", "--locale", locale)
			Eventually(session).Should(Exit(0))

			underscored_locale := strings.Replace(locale, "-", "_", -1)
			session = helpers.CF("config", "--locale", underscored_locale)
			Eventually(session).Should(Exit(0))
		},

		Entry("Chinese (Simplified)", "zh-Hans"),
		Entry("Chinese (Traditional)", "zh-Hant"),
		Entry("English (United States)", "en-US"),
		Entry("French", "fr-FR"),
		Entry("German", "de-DE"),
		Entry("Italian", "it-IT"),
		Entry("Japanese", "ja-JP"),
		Entry("Korean", "ko-KR"),
		Entry("Portuguese (Brazil)", "pt-BR"),
		Entry("Spanish", "es-ES"),
		Entry("CLEAR", "CLEAR"),
	)

	When("the locale provided is not a valid option", func() {
		It("fails with the appropriate errors", func() {
			session := helpers.CF("config", "--locale", "bora bora")
			Eventually(session.Err).Should(Say("Incorrect Usage: LOCALE must be CLEAR, de-DE, en-US, es-ES, fr-FR, it-IT, ja-JP, ko-KR, pt-BR, zh-Hans, zh-Hant"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	DescribeTable("allows setting trace to",
		func(trace string) {
			session := helpers.CF("config", "--trace", trace)
			Eventually(session).Should(Exit(0))
		},

		Entry("true", "true"),
		Entry("false", "false"),
		Entry("path", "some/path"),
	)
})
