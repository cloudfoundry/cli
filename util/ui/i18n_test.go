package ui_test

import (
	. "code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/util/ui/uifakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("i18n", func() {
	var fakeConfig *uifakes.FakeConfig

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
	})

	Describe("GetTranslationFunc", func() {
		DescribeTable("defaults to english",
			func(locale string) {
				fakeConfig.LocaleReturns(locale)

				translationFunc, err := GetTranslationFunc(fakeConfig)
				Expect(err).ToNot(HaveOccurred())

				Expect(translationFunc("\nApp started\n")).To(Equal("\nApp started\n"))
			},

			Entry("when left blank", ""),
			Entry("when given gibberish", "asdfasfsadfsadfsadfa"),
			Entry("when given an unsupported language", "pt-PT"),
		)

		Context("when the config file is set", func() {
			DescribeTable("returns the correct language translationFunc",
				func(locale string, expectedTranslation string) {
					fakeConfig.LocaleReturns(locale)

					translationFunc, err := GetTranslationFunc(fakeConfig)
					Expect(err).ToNot(HaveOccurred())

					Expect(translationFunc("\nApp started\n")).To(Equal(expectedTranslation))
				},

				Entry("German", "de-DE", "\nApp gestartet\n"),
				Entry("English", "en-US", "\nApp started\n"),
				Entry("Spanish", "es-ES", "\nApp iniciada\n"),
				Entry("French", "fr-FR", "\nApplication démarrée\n"),
				Entry("Italian", "it-IT", "\nApplicazione avviata\n"),
				Entry("Japanese", "ja-JP", "\nアプリが開始されました\n"),
				Entry("Korean", "ko-KR", "\n앱 시작됨\n"),
				Entry("Brazilian Portuguese", "pt-BR", "\nApp iniciado\n"),
				Entry("Chinese (Simplified)", "zh-HANS", "\n应用程序已启动\n"),
				Entry("Chinese (Traditional)", "zh-HANT", "\n已啟動應用程式\n"),

				// The following locales use the zh-hant translations
				Entry("Chinese (Traditional and using Taiwanese terms)", "zh-TW", "\n已啟動應用程式\n"),
				Entry("Chinese (Traditional and using Hong Kong terms)", "zh-HK", "\n已啟動應用程式\n"),
			)

			Context("when provided keys to iterpolate", func() {
				BeforeEach(func() {
					fakeConfig.LocaleReturns("fr-FR")
				})

				It("interpolates them properly", func() {
					translationFunc, err := GetTranslationFunc(fakeConfig)
					Expect(err).ToNot(HaveOccurred())
					translated := translationFunc("\nApp {{.AppName}} was started using this command `{{.Command}}`\n", map[string]interface{}{
						"AppName": "some-app-name",
						"Command": "some-command-name",
					})
					Expect(translated).To(Equal("\nL'application some-app-name a été démarrée avec la commande `some-command-name`\n"))
				})
			})
		})

		Context("when the translation does not have a value", func() {
			It("uses the id for the translation", func() {
				translationFunc, err := GetTranslationFunc(fakeConfig)
				Expect(err).ToNot(HaveOccurred())
				translated := translationFunc("api version:")
				Expect(translated).To(Equal("api version:"))
			})
		})
	})

	Describe("ParseLocale", func() {
		DescribeTable("returns the correct language translationFunc",
			func(locale string, expectedLocale string) {
				Expect(ParseLocale(locale)).To(Equal(expectedLocale))
			},

			Entry("German with underscore and caps", "DE_DE", "de-de"),
			Entry("German", "de-DE", "de-de"),
			Entry("English", "en-US", "en-us"),
			Entry("Spanish", "es-ES", "es-es"),
			Entry("French", "fr-FR", "fr-fr"),
			Entry("Italian", "it-IT", "it-it"),
			Entry("Japanese", "ja-JP", "ja-jp"),
			Entry("Korean", "ko-KR", "ko-kr"),
			Entry("Brazilian Portuguese", "pt-BR", "pt-br"),
			Entry("Chinese (Simplified)", "zh-HANS", "zh-hans"),
			Entry("Chinese (Traditional)", "zh-HANT", "zh-hant"),

			// The following locales use the zh-hant translations
			Entry("Chinese (Traditional and using Taiwanese terms)", "zh-TW", "zh-hant"),
			Entry("Chinese (Traditional and using Hong Kong terms)", "zh-HK", "zh-hant"),
		)
	})
})
