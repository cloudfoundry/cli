package ui_test

import (
	. "code.cloudfoundry.org/cli/commands/ui"
	"code.cloudfoundry.org/cli/commands/ui/uifakes"
	"code.cloudfoundry.org/cli/utils/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("UI", func() {
	var (
		ui         UI
		fakeConfig *uifakes.FakeConfig
	)

	// type TranslateFunc func(translationID string, args ...interface{}) string
	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(config.ColorEnabled)

		var err error
		ui, err = NewUI(fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		ui.Out = NewBuffer()
	})

	Describe("DisplayText", func() {
		Context("when only a string is passed in", func() {
			It("displays the string to Out with a newline", func() {
				ui.DisplayText("some-string")

				Expect(ui.Out).To(Say("some-string\n"))
			})
		})

		Context("when a map is passed in", func() {
			It("merges the map content with the string", func() {
				ui.DisplayText("some-string {{.SomeMapValue}}", map[string]interface{}{
					"SomeMapValue": "my-map-value",
				})

				Expect(ui.Out).To(Say("some-string my-map-value\n"))
			})
		})

		Context("when multiple maps are passed in", func() {
			It("merges all map value with the string", func() {
				ui.DisplayText("some-string {{.SomeMapValue}} {{.SomeOtherMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "my-map-value",
					},
					map[string]interface{}{
						"SomeOtherMapValue": "my-other-map-value",
					},
				)

				Expect(ui.Out).To(Say("some-string my-map-value my-other-map-value\n"))
			})

			Context("when maps share the same key", func() {
				It("keeps the rightmost map value", func() {
					ui.DisplayText("some-string {{.SomeMapValue}}",
						map[string]interface{}{
							"SomeMapValue": "my-map-value",
						},
						map[string]interface{}{
							"SomeMapValue": "my-other-map-value",
						},
					)

					Expect(ui.Out).To(Say("some-string my-other-map-value\n"))
				})
			})

			Context("when the local is not set to 'en-us'", func() {
				BeforeEach(func() {
					fakeConfig = new(uifakes.FakeConfig)
					fakeConfig.ColorEnabledReturns(config.ColorEnabled)
					fakeConfig.LocaleReturns("fr-FR")

					var err error
					ui, err = NewUI(fakeConfig)
					Expect(err).NotTo(HaveOccurred())

					ui.Out = NewBuffer()
				})

				It("translates the main string passed to DisplayText", func() {
					ui.DisplayText("\nTIP: Use '{{.Command}}' to target new org",
						map[string]interface{}{
							"Command": "foo",
						},
					)

					Expect(ui.Out).To(Say("\nASTUCE : utilisez 'foo' pour cibler une nouvelle organisation"))
				})

				It("translates the main string and keys passed to DisplayTextWithKeyTranslations", func() {
					ui.DisplayTextWithKeyTranslations("   {{.CommandName}} - {{.CommandDescription}}",
						[]string{"CommandDescription"},
						map[string]interface{}{
							"CommandName":        "ADVANCED", // In translation file, should not be translated
							"CommandDescription": "A command line tool to interact with Cloud Foundry",
						})

					Expect(ui.Out).To(Say("   ADVANCED - Outil de ligne de commande permettant d'interagir avec Cloud Foundry"))
				})
			})
		})
	})

	Describe("DisplayTextWithKeyTranslations", func() {
		Context("when the local is not set to 'en-us'", func() {
			BeforeEach(func() {
				fakeConfig = new(uifakes.FakeConfig)
				fakeConfig.ColorEnabledReturns(config.ColorEnabled)
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = NewBuffer()
			})

			It("translates the main string and keys passed to DisplayTextWithKeyTranslations", func() {
				ui.DisplayTextWithKeyTranslations("   {{.CommandName}} - {{.CommandDescription}}",
					[]string{"CommandDescription"},
					map[string]interface{}{
						"CommandName":        "ADVANCED", // In translation file, should not be translated
						"CommandDescription": "A command line tool to interact with Cloud Foundry",
					})

				Expect(ui.Out).To(Say("   ADVANCED - Outil de ligne de commande permettant d'interagir avec Cloud Foundry"))
			})
		})
	})

	Describe("DisplayNewline", func() {
		It("displays a new line", func() {
			ui.DisplayNewline()

			Expect(ui.Out).To(Say("\n"))
		})
	})

	Describe("DisplayHelpHeader", func() {
		It("bolds and colorizes the input string", func() {
			ui.DisplayHelpHeader("some-text")
			Expect(ui.Out).To(Say("\x1b\\[38;1msome-text\x1b\\[0m"))
		})

		Context("when the local is not set to 'en-us'", func() {
			BeforeEach(func() {
				fakeConfig = new(uifakes.FakeConfig)
				fakeConfig.ColorEnabledReturns(config.ColorEnabled)
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = NewBuffer()
			})

			It("bolds and colorizes the input string", func() {
				ui.DisplayHelpHeader("FEATURE FLAGS")
				Expect(ui.Out).To(Say("\x1b\\[38;1mINDICATEURS DE FONCTION\x1b\\[0m"))
			})
		})
	})
})
