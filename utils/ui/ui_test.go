package ui_test

import (
	"errors"

	"code.cloudfoundry.org/cli/utils/configv3"
	. "code.cloudfoundry.org/cli/utils/ui"
	"code.cloudfoundry.org/cli/utils/ui/uifakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("UI", func() {
	var (
		ui         *UI
		fakeConfig *uifakes.FakeConfig
		inBuffer   *Buffer
	)

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)

		var err error
		ui, err = NewUI(fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		inBuffer = NewBuffer()
		ui.In = inBuffer
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
					fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)
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
				fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)
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
				fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)
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

	Describe("DisplayBoolPrompt", func() {
		It("displays the prompt", func() {
			ui.DisplayBoolPrompt("some-prompt", false)
			Expect(ui.Out).To(Say("some-prompt\x1b\\[36;1m>>\x1b\\[0m"))
		})

		Context("when the interact library returns an error", func() {
			It("returns the error", func() {
				inBuffer.Write([]byte("invalid\n"))
				_, err := ui.DisplayBoolPrompt("some-prompt", false)
				Expect(err).To(HaveOccurred())
			})
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

			It("returns true", func() {
				response, err := ui.DisplayBoolPrompt("some-prompt", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(BeFalse())
			})
		})

		Context("when the user chooses the default", func() {
			BeforeEach(func() {
				inBuffer.Write([]byte("\n"))
			})

			Context("when the default is false", func() {
				It("return false", func() {
					response, err := ui.DisplayBoolPrompt("some-prompt", false)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeFalse())
				})
			})

			Context("when the default is true", func() {
				It("returns true", func() {
					response, err := ui.DisplayBoolPrompt("some-prompt", true)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeTrue())
				})
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
				fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)
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

	Describe("DisplayError", func() {
		Context("when passed a TranslatableError", func() {
			var fakeTranslateErr *uifakes.FakeTranslatableError

			BeforeEach(func() {
				fakeTranslateErr = new(uifakes.FakeTranslatableError)
				fakeTranslateErr.TranslateReturns("I am an error")

				ui.DisplayError(fakeTranslateErr)
			})

			It("displays the error to Err and displays the FAILED text in red to Out", func() {
				Expect(ui.Err).To(Say("I am an error\n"))
				Expect(ui.Out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
			})

			Context("when the locale is not set to 'en-us'", func() {
				It("translates the error text and the FAILED text", func() {
					Expect(fakeTranslateErr.TranslateCallCount()).To(Equal(1))
					Expect(fakeTranslateErr.TranslateArgsForCall(0)).NotTo(BeNil())
				})
			})
		})

		Context("when passed a generic error", func() {
			var err error

			BeforeEach(func() {
				err = errors.New("I am a BANANA!")
				ui.DisplayError(err)
			})

			It("displays the error to Err and displays the FAILED text in red to Out", func() {
				Expect(ui.Err).To(Say("I am a BANANA!\n"))
				Expect(ui.Out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
			})
		})
	})

	Describe("DisplayWarning", func() {
		It("displays the warning", func() {
			ui.DisplayWarning("some template string with value = {{.SomeKey}}", map[string]interface{}{
				"SomeKey": "some-value",
			})

			Expect(ui.Err).To(Say("some template string with value = some-value"))
		})

		Context("when the locale is not set to en-US", func() {
			BeforeEach(func() {
				fakeConfig = new(uifakes.FakeConfig)
				fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				ui.Out = NewBuffer()
				ui.Err = NewBuffer()
			})

			It("displays the translated warning", func() {
				ui.DisplayWarning("'{{.VersionShort}}' and '{{.VersionLong}}' are also accepted.", map[string]interface{}{
					"VersionShort": "some-value",
					"VersionLong":  "some-other-value",
				})

				Expect(ui.Err).To(Say("'some-value' et 'some-other-value' sont également acceptés.\n"))
			})
		})
	})

	Describe("DisplayWarnings", func() {
		It("displays the warnings", func() {
			ui.DisplayWarnings([]string{"warnings-1", "warnings-2"})

			Expect(ui.Err).To(Say("warnings-1\n"))
			Expect(ui.Err).To(Say("warnings-2\n"))
		})

		Context("when the locale is not set to en-US", func() {
			BeforeEach(func() {
				fakeConfig = new(uifakes.FakeConfig)
				fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				ui.Out = NewBuffer()
				ui.Err = NewBuffer()
			})

			It("displays the translated warnings", func() {
				ui.DisplayWarnings([]string{"warnings-1", "FEATURE FLAGS"})

				Expect(ui.Err).To(Say("warnings-1\n"))
				Expect(ui.Err).To(Say("INDICATEURS DE FONCTION\n"))
			})
		})
	})
})
