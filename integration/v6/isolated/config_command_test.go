package isolated

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("config command", func() {
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
	)
})
