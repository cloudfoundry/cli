package i18n_test

import (
	"code.cloudfoundry.org/cli/cf/i18n"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("I18n", func() {
	Describe("SupportedLocales", func() {
		It("returns the list of locales in resources", func() {
			supportedLocales := i18n.SupportedLocales()
			Expect(supportedLocales).To(ContainElement("fr-FR"))
			Expect(supportedLocales).To(ContainElement("it-IT"))
			Expect(supportedLocales).To(ContainElement("ja-JP"))
			Expect(supportedLocales).To(ContainElement("zh-Hans"))
			Expect(supportedLocales).To(ContainElement("zh-Hant"))
			Expect(supportedLocales).To(ContainElement("en-US"))
			Expect(supportedLocales).To(ContainElement("es-ES"))
			Expect(supportedLocales).To(ContainElement("pt-BR"))
			Expect(supportedLocales).To(ContainElement("de-DE"))
			Expect(supportedLocales).To(ContainElement("ko-KR"))
		})
	})

	Describe("IsSupportedLocale", func() {
		It("returns true for supported locales", func() {
			Expect(i18n.IsSupportedLocale("fr-FR")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("Fr_Fr")).To(BeTrue())

			Expect(i18n.IsSupportedLocale("it-IT")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("ja-JP")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("zh-Hans")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("zh-Hant")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("en-US")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("es-ES")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("pt-BR")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("de-DE")).To(BeTrue())
			Expect(i18n.IsSupportedLocale("ko-KR")).To(BeTrue())
		})

		It("returns false for unsupported locales", func() {
			Expect(i18n.IsSupportedLocale("potato-Tomato")).To(BeFalse())
		})
	})
})
