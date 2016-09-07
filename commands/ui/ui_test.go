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

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(config.ColorEnabled)

		var err error
		ui, err = NewUI(fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		ui.Out = NewBuffer()
		ui.Err = NewBuffer()
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

	Describe("DisplayPair", func() {
		Context("when the local is 'en-us'", func() {
			It("prints out the key and value", func() {
				ui.DisplayPair("some-key", "App {{.AppName}} does not exist.",
					map[string]interface{}{
						"AppName": "some-app-name",
					})
				Expect(ui.Out).To(Say("some-key: App some-app-name does not exist.\n"))
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

			It("prints out the key and value", func() {
				ui.DisplayPair("ADVANCED", "App {{.AppName}} does not exist.",
					map[string]interface{}{
						"AppName": "some-app-name",
					})
				Expect(ui.Out).To(Say("AVANCE: L'application some-app-name n'existe pas.\n"))
			})
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

	Describe("DisplayHeaderFlavorText", func() {
		It("displays the header with cyan subject values", func() {
			ui.DisplayHeaderFlavorText("some text {{.Key}}",
				map[string]interface{}{
					"Key": "Value",
				})
			Expect(ui.Out).To(Say("some text \x1b\\[36;1mValue\x1b\\[0m"))
		})
	})

	Describe("DisplayOK", func() {
		It("displays the OK text in green", func() {
			ui.DisplayOK()
			Expect(ui.Out).To(Say("\x1b\\[32;1mOK\x1b\\[0m"))
		})
	})

	Describe("DisplayErrorMessage", func() {
		Context("when only a string is passed in", func() {
			It("displays the string to Err and outputs FAILED to Out", func() {
				ui.DisplayErrorMessage("some-string")

				Expect(ui.Err).To(Say("some-string\n"))
				Expect(ui.Out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
			})
		})

		Context("when a map is passed in", func() {
			It("merges the map content with the string", func() {
				ui.DisplayErrorMessage("some-string {{.SomeMapValue}}", map[string]interface{}{
					"SomeMapValue": "my-map-value",
				})

				Expect(ui.Err).To(Say("some-string my-map-value\n"))
				Expect(ui.Out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
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
					ui.Err = NewBuffer()
				})

				It("translates the main string that gets passed to err and outputs a translated FAILED", func() {
					ui.DisplayErrorMessage("\nTIP: Use '{{.Command}}' to target new org",
						map[string]interface{}{
							"Command": "foo",
						},
					)

					Expect(ui.Err).To(Say("\nASTUCE : utilisez 'foo' pour cibler une nouvelle organisation"))
					Expect(ui.Out).To(Say("\x1b\\[31;1mECHEC\x1b\\[0m\n"))
				})
			})
		})
	})

	Describe("DisplayError", func() {
		var fakeTranslateErr *uifakes.FakeTranslatableError

		BeforeEach(func() {
			fakeTranslateErr = new(uifakes.FakeTranslatableError)
			fakeTranslateErr.SetTranslationReturns(fakeTranslateErr)
			fakeTranslateErr.ErrorReturns("I am an error")

			ui.DisplayError(fakeTranslateErr)
		})

		It("displays the error to Err and displays the FAILED text in red to Out", func() {
			Expect(ui.Err).To(Say("I am an error\n"))
			Expect(ui.Out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
		})

		Context("when the locale is not set to 'en-us'", func() {
			It("translates the error text and the FAILED text", func() {
				Expect(fakeTranslateErr.SetTranslationCallCount()).To(Equal(1))
				Expect(fakeTranslateErr.SetTranslationArgsForCall(0)).NotTo(BeNil())
			})
		})
	})
})
