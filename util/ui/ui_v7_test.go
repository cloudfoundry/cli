package ui_test

import (
	"code.cloudfoundry.org/cli/util/configv3"
	. "code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/util/ui/uifakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("UI", func() {
	var (
		ui         *UI
		fakeConfig *uifakes.FakeConfig
		out        *Buffer
		errBuff    *Buffer
	)

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)

		var err error
		ui, err = NewUI(fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		out = NewBuffer()
		ui.Out = out
		ui.OutForInteraction = out
		errBuff = NewBuffer()
		ui.Err = errBuff
	})

	// Covers the happy paths, additional cases are tested in TranslateText
	Describe("DisplayWarnings", func() {
		It("displays the warnings to ui.Err", func() {
			ui.DisplayWarnings([]string{"warning-1", "warning-2"})
			Expect(ui.Err).To(Say("warning-1\n"))
			Expect(ui.Err).To(Say("warning-2\n"))
		})

		When("the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				errBuff = NewBuffer()
				ui.Err = errBuff
			})

			When("there are multiple warnings", func() {
				It("displays translated warnings to ui.Err", func() {
					ui.DisplayWarnings([]string{"un-translatable warning", "FEATURE FLAGS", "Number of instances"})
					Expect(string(errBuff.Contents())).To(Equal("un-translatable warning\nINDICATEURS DE FONCTION\nNombre d'instances\n"))
				})
			})

			When("there is a single warning ", func() {
				It("displays the translated warning to ui.Err", func() {
					ui.DisplayWarnings([]string{"un-translatable warning"})
					Expect(string(errBuff.Contents())).To(Equal("un-translatable warning\n"))
				})
			})

			Context("there are no warnings", func() {
				It("does not print out a new line", func() {
					ui.DisplayWarnings(nil)
					Expect(errBuff.Contents()).To(BeEmpty())
				})
			})
		})
	})
})
