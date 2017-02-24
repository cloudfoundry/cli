package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Locale", func() {
	var locale Locale

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := locale.Complete(prefix)
				Expect(completions).To(ConsistOf(matches))
			},
			Entry("completes to 'en-US' and 'es-ES' when passed 'e'", "e",
				[]flags.Completion{{Item: "es-ES"}, {Item: "en-US"}}),
			Entry("completes to 'en-US' when passed 'en_'", "en_",
				[]flags.Completion{{Item: "en-US"}}),
			Entry("completes to 'en-US' when passed 'eN_'", "eN_",
				[]flags.Completion{{Item: "en-US"}}),
			Entry("returns CLEAR, de-DE, en-US, es-ES, fr-FR, it-IT, ja-JP, ko-KR, pt-BR, zh-Hans, zh-Hant when passed nothing", "",
				[]flags.Completion{{Item: "CLEAR"}, {Item: "de-DE"}, {Item: "en-US"}, {Item: "es-ES"}, {Item: "fr-FR"}, {Item: "it-IT"}, {Item: "ja-JP"}, {Item: "ko-KR"}, {Item: "pt-BR"}, {Item: "zh-Hans"}, {Item: "zh-Hant"}}),
			Entry("completes to nothing when passed 'wut'", "wut",
				[]flags.Completion{}),
		)
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			locale = Locale{}
		})

		It("accepts en-us", func() {
			err := locale.UnmarshalFlag("en-us")
			Expect(err).ToNot(HaveOccurred())
			Expect(locale.Locale).To(Equal("en-US"))
		})

		It("accepts en_us", func() {
			err := locale.UnmarshalFlag("en_us")
			Expect(err).ToNot(HaveOccurred())
			Expect(locale.Locale).To(Equal("en-US"))
		})

		It("accepts ja-jp", func() {
			err := locale.UnmarshalFlag("ja-jp")
			Expect(err).ToNot(HaveOccurred())
			Expect(locale.Locale).To(Equal("ja-JP"))
		})

		It("errors on anything else", func() {
			err := locale.UnmarshalFlag("I AM A BANANANANANANANANAE")
			Expect(err).To(MatchError(&flags.Error{
				Type:    flags.ErrRequired,
				Message: `LOCALE must be CLEAR, de-DE, en-US, es-ES, fr-FR, it-IT, ja-JP, ko-KR, pt-BR, zh-Hans, zh-Hant`,
			}))
			Expect(locale.Locale).To(BeEmpty())
		})
	})
})
