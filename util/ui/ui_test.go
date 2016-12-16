package ui_test

import (
	"errors"

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
	)

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)

		var err error
		ui, err = NewUI(fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		ui.Out = NewBuffer()
		ui.Err = NewBuffer()
	})

	Describe("TranslateText", func() {
		It("returns the template", func() {
			Expect(ui.TranslateText("some-template")).To(Equal("some-template"))
		})

		Context("when an optional map is passed in", func() {
			It("returns the template with map values substituted in", func() {
				expected := ui.TranslateText(
					"template {{.SomeMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					})
				Expect(expected).To(Equal("template map-value"))
			})
		})

		Context("when multiple optional maps are passed in", func() {
			It("returns the template with only the first map values substituted in", func() {
				expected := ui.TranslateText(
					"template with {{.SomeMapValue}} and {{.SomeOtherMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					},
					map[string]interface{}{
						"SomeOtherMapValue": "other-map-value",
					})
				Expect(expected).To(Equal("template with map-value and <no value>"))
			})
		})

		Context("when the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the translated template", func() {
				expected := ui.TranslateText("   View allowable quotas with 'CF_NAME quotas'")
				Expect(expected).To(Equal("   Affichez les quotas pouvant être alloués avec 'CF_NAME quotas'"))
			})
		})
	})

	Describe("DisplayOK", func() {
		It("displays 'OK' in green and bold", func() {
			ui.DisplayOK()
			Expect(ui.Out).To(Say("\x1b\\[32;1mOK\x1b\\[0m"))
		})
	})

	Describe("DisplayNewline", func() {
		It("displays a new line", func() {
			ui.DisplayNewline()
			Expect(ui.Out).To(Say("\n"))
		})
	})

	Describe("DisplayBoolPrompt", func() {
		var inBuffer *Buffer

		BeforeEach(func() {
			inBuffer = NewBuffer()
			ui.In = inBuffer
		})

		It("displays the passed in string", func() {
			ui.DisplayBoolPrompt("some-prompt", false)
			Expect(ui.Out).To(Say("some-prompt\x1b\\[36;1m>>\x1b\\[0m"))
		})

		Context("when the user chooses yes", func() {
			BeforeEach(func() {
				inBuffer.Write([]byte("y\n"))
			})

			It("returns true", func() {
				response, err := ui.DisplayBoolPrompt("some-prompt", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(BeTrue())
			})
		})

		Context("when the user chooses no", func() {
			BeforeEach(func() {
				inBuffer.Write([]byte("n\n"))
			})

			It("returns false", func() {
				response, err := ui.DisplayBoolPrompt("some-prompt", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(BeFalse())
			})
		})

		Context("when the user chooses the default", func() {
			BeforeEach(func() {
				inBuffer.Write([]byte("\n"))
			})

			Context("when the default is true", func() {
				It("returns true", func() {
					response, err := ui.DisplayBoolPrompt("some-prompt", true)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeTrue())
				})
			})

			Context("when the default is false", func() {
				It("returns false", func() {
					response, err := ui.DisplayBoolPrompt("some-prompt", false)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeFalse())
				})
			})
		})

		Context("when the interact library returns an error", func() {
			It("returns the error", func() {
				inBuffer.Write([]byte("invalid\n"))
				_, err := ui.DisplayBoolPrompt("some-prompt", false)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DisplayTable", func() {
		It("displays a string matrix as a table with the provided prefix and padding to ui.Out", func() {
			ui.DisplayTable(
				"some-prefix",
				[][]string{
					{"aaaaaaaaa", "bb", "ccccccc"},
					{"dddd", "eeeeeeeeeee", "fff"},
					{"gg", "hh", "ii"},
				},
				3)
			Expect(ui.Out).To(Say(`some-prefixaaaaaaaaa   bb            ccccccc
some-prefixdddd        eeeeeeeeeee   fff
some-prefixgg          hh            ii`))
		})
	})

	// Covers the happy paths, additional cases are tested in TranslateText.
	Describe("DisplayText", func() {
		It("displays the template with map values substituted in to ui.Out with a newline", func() {
			ui.DisplayText(
				"template with {{.SomeMapValue}}",
				map[string]interface{}{
					"SomeMapValue": "map-value",
				})
			Expect(ui.Out).To(Say("template with map-value\n"))
		})

		Context("when the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = NewBuffer()
			})

			It("displays the translated template with map values substituted in to ui.Out", func() {
				ui.DisplayText(
					"\nTIP: Use '{{.Command}}' to target new org",
					map[string]interface{}{
						"Command": "foo",
					})
				Expect(ui.Out).To(Say("\nASTUCE : utilisez 'foo' pour cibler une nouvelle organisation"))
			})
		})
	})

	// Covers the happy paths, additional cases are tested in TranslateText.
	Describe("DisplayPair", func() {
		It("displays the pair with map values substituted in to ui.Out", func() {
			ui.DisplayPair(
				"some-key",
				"some-value with {{.SomeMapValue}}",
				map[string]interface{}{
					"SomeMapValue": "map-value",
				})
			Expect(ui.Out).To(Say("some-key: some-value with map-value\n"))
		})

		Context("when the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = NewBuffer()
			})

			It("displays the translated pair with map values substituted in to ui.Out", func() {
				ui.DisplayPair(
					"ADVANCED",
					"App {{.AppName}} does not exist.",
					map[string]interface{}{
						"AppName": "some-app-name",
					})
				Expect(ui.Out).To(Say("AVANCE: L'application some-app-name n'existe pas.\n"))
			})
		})
	})

	Describe("DisplayHeader", func() {
		It("displays the header colorized and bolded to ui.Out", func() {
			ui.DisplayHeader("some-header")
			Expect(ui.Out).To(Say("\x1b\\[38;1msome-header\x1b\\[0m"))
		})

		Context("when the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = NewBuffer()
			})

			It("displays the translated header colorized and bolded to ui.Out", func() {
				ui.DisplayHeader("FEATURE FLAGS")
				Expect(ui.Out).To(Say("\x1b\\[38;1mINDICATEURS DE FONCTION\x1b\\[0m"))
			})
		})
	})

	Describe("DisplayTextWithFlavor", func() {
		It("displays the template to ui.Out", func() {
			ui.DisplayTextWithFlavor("some-template")
			Expect(ui.Out).To(Say("some-template"))
		})

		Context("when an optional map is passed in", func() {
			It("displays the template with map values colorized, bolded, and substituted in to ui.Out", func() {
				ui.DisplayTextWithFlavor(
					"template with {{.SomeMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					})
				Expect(ui.Out).To(Say("template with \x1b\\[36;1mmap-value\x1b\\[0m"))
			})
		})

		Context("when multiple optional maps are passed in", func() {
			It("displays the template with only the first map values colorized, bolded, and substituted in to ui.Out", func() {
				ui.DisplayTextWithFlavor(
					"template with {{.SomeMapValue}} and {{.SomeOtherMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					},
					map[string]interface{}{
						"SomeOtherMapValue": "other-map-value",
					})
				Expect(ui.Out).To(Say("template with \x1b\\[36;1mmap-value\x1b\\[0m and <no value>"))
			})
		})

		Context("when the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = NewBuffer()
			})

			It("displays the translated template with map values colorized, bolded and substituted in to ui.Out", func() {
				ui.DisplayTextWithFlavor(
					"App {{.AppName}} does not exist.",
					map[string]interface{}{
						"AppName": "some-app-name",
					})
				Expect(ui.Out).To(Say("L'application \x1b\\[36;1msome-app-name\x1b\\[0m n'existe pas.\n"))
			})
		})
	})

	// Covers the happy paths, additional cases are tested in TranslateText.
	Describe("DisplayWarning", func() {
		It("displays the warning to ui.Err", func() {
			ui.DisplayWarning(
				"template with {{.SomeMapValue}}",
				map[string]interface{}{
					"SomeMapValue": "map-value",
				})
			Expect(ui.Err).To(Say("template with map-value"))
		})

		Context("when the locale is not set to english", func() {
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

	// Covers the happy paths, additional cases are tested in TranslateText.
	Describe("DisplayWarnings", func() {
		It("displays the warnings to ui.Err", func() {
			ui.DisplayWarnings([]string{"warning-1", "warning-2"})
			Expect(ui.Err).To(Say("warning-1\n"))
			Expect(ui.Err).To(Say("warning-2\n"))
		})

		Context("when the locale is not set to english", func() {
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
			})
		})
	})

	Describe("DisplayError", func() {
		Context("when passed a TranslatableError", func() {
			var fakeTranslateErr *uifakes.FakeTranslatableError

			BeforeEach(func() {
				fakeTranslateErr = new(uifakes.FakeTranslatableError)
				fakeTranslateErr.TranslateReturns("I am an error")

				ui.DisplayError(fakeTranslateErr)
			})

			It("displays the error to ui.Err and displays FAILED in bold red to ui.Out", func() {
				Expect(ui.Err).To(Say("I am an error\n"))
				Expect(ui.Out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
			})

			Context("when the locale is not set to english", func() {
				It("translates the error text", func() {
					Expect(fakeTranslateErr.TranslateCallCount()).To(Equal(1))
					Expect(fakeTranslateErr.TranslateArgsForCall(0)).NotTo(BeNil())
				})
			})
		})

		Context("when passed a generic error", func() {
			It("displays the error text to ui.Err and displays FAILED in bold red to ui.Out", func() {
				ui.DisplayError(errors.New("I am a BANANA!"))
				Expect(ui.Err).To(Say("I am a BANANA!\n"))
				Expect(ui.Out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
			})
		})
	})
})
