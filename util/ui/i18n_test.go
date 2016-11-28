package ui_test

import (
	. "code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/util/ui/uifakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("i18n", func() {
	var fakeConfig *uifakes.FakeConfig

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
	})

	Context("when the config file is empty", func() {
		It("returns back default translation", func() {
			translationFunc, err := GetTranslationFunc(fakeConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(translationFunc("\nApp started\n")).To(Equal("\nApp started\n"))
		})
	})

	Context("when the config file is set", func() {
		Context("when we support the language", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")
			})

			It("returns back default translation", func() {
				translationFunc, err := GetTranslationFunc(fakeConfig)
				Expect(err).ToNot(HaveOccurred())
				Expect(translationFunc("\nApp started\n")).To(Equal("\nApplication démarrée\n"))
			})
		})
	})
})
