// +build !V7

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
		ui.OutForInteration = out
		errBuff = NewBuffer()
		ui.Err = errBuff
	})

	// Covers the happy paths, additional cases are tested in TranslateText
	Describe("DisplayWarning", func() {
		It("displays the warning to ui.Err", func() {
			ui.DisplayWarning(
				"template with {{.SomeMapValue}}",
				map[string]interface{}{
					"SomeMapValue": "map-value",
				})
			Expect(ui.Err).To(Say("template with map-value\n\n"))
		})

		When("the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Err = NewBuffer()
			})

			It("displays the translated warning to ui.Err", func() {
				ui.DisplayWarning(
					"'{{.VersionShort}}' and '{{.VersionLong}}' are also accepted.",
					map[string]interface{}{
						"VersionShort": "some-value",
						"VersionLong":  "some-other-value",
					})
				Expect(ui.Err).To(Say("'some-value' et 'some-other-value' sont également acceptés.\n"))
			})
		})
	})

	// Covers the happy paths, additional cases are tested in TranslateText
	Describe("DisplayWarnings", func() {
		It("displays the warnings to ui.Err", func() {
			ui.DisplayWarnings([]string{"warning-1", "warning-2"})
			Expect(ui.Err).To(Say("warning-1\n"))
			Expect(ui.Err).To(Say("warning-2\n"))
			Expect(ui.Err).To(Say("\n"))
		})

		When("the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Err = NewBuffer()
			})

			It("displays the translated warnings to ui.Err", func() {
				ui.DisplayWarnings([]string{"Also delete any mapped routes", "FEATURE FLAGS"})
				Expect(ui.Err).To(Say("Supprimer aussi les routes mappées\n"))
				Expect(ui.Err).To(Say("INDICATEURS DE FONCTION\n"))
				Expect(ui.Err).To(Say("\n"))
			})
		})

		Context("does not display newline when warnings are empty", func() {
			It("does not print out a new line", func() {
				ui.DisplayWarnings(nil)
				Expect(errBuff.Contents()).To(BeEmpty())
			})
		})
	})
})
