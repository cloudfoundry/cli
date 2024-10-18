package ui_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/v8/command/translatableerror/translatableerrorfakes"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	. "code.cloudfoundry.org/cli/v8/util/ui"
	"code.cloudfoundry.org/cli/v8/util/ui/uifakes"
	. "github.com/onsi/ginkgo/v2"
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

	Describe("DisplayDeprecationWarning", func() {
		It("displays the deprecation warning to ui.Err", func() {
			ui.DisplayDeprecationWarning()
			Expect(ui.Err).To(Say("Deprecation warning: This command has been deprecated. This feature will be removed in the future.\n"))
		})

		When("the locale is not set to English", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Err = NewBuffer()
			})

			PIt("displays the translated deprecation warning to ui.Err", func() {
				// TODO: Test implementation awaits translated version of deprecation warning string literal #164098152.
			})
		})
	})

	Describe("DisplayFileDeprecationWarning", func() {
		It("displays the `cf files` deprecation warning to ui.Err", func() {
			ui.DisplayFileDeprecationWarning()
			Expect(ui.Err).To(Say("Deprecation warning: This command has been deprecated and will be removed in the future. For similar functionality, please use the `cf ssh` command instead.\n"))
		})

		When("the locale is not set to English", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Err = NewBuffer()
			})

			PIt("displays the translated deprecation warning to ui.Err", func() {
				// TODO: Test implementation awaits translated version of deprecation warning string literal #164098103.
			})
		})
	})

	Describe("DisplayError", func() {
		When("passed a TranslatableError", func() {
			var fakeTranslateErr *translatableerrorfakes.FakeTranslatableError

			BeforeEach(func() {
				fakeTranslateErr = new(translatableerrorfakes.FakeTranslatableError)
				fakeTranslateErr.TranslateReturns("I am an error")

				ui.DisplayError(fakeTranslateErr)
			})

			It("displays the error to ui.Err and displays FAILED in bold red to ui.Out", func() {
				Expect(ui.Err).To(Say("I am an error\n"))
				Expect(out).To(Say("\u001B\\[31;1mFAILED\u001B\\[0;22m\n"))
			})

			When("the locale is not set to english", func() {
				It("translates the error text", func() {
					Expect(fakeTranslateErr.TranslateCallCount()).To(Equal(1))
					Expect(fakeTranslateErr.TranslateArgsForCall(0)).NotTo(BeNil())
				})
			})
		})

		When("passed a generic error", func() {
			It("displays the error text to ui.Err and displays FAILED in bold red to ui.Out", func() {
				ui.DisplayError(errors.New("I am a BANANA!"))
				Expect(ui.Err).To(Say("I am a BANANA!\n"))
				Expect(out).To(Say("\u001B\\[31;1mFAILED\u001B\\[0;22m\n"))
			})
		})
	})

	Describe("DisplayHeader", func() {
		It("displays the header colorized and bolded to ui.Out", func() {
			ui.DisplayHeader("some-header")
			Expect(out).To(Say("\u001B\\[1msome-header\u001B\\[22m"))
		})

		When("the locale is not set to English", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("displays the translated header colorized and bolded to ui.Out", func() {
				ui.DisplayHeader("FEATURE FLAGS")
				Expect(out).To(Say("\u001B\\[1mINDICATEURS DE FONCTION\u001B\\[22m"))
			})
		})
	})

	Describe("DisplayNewline", func() {
		It("displays a new line", func() {
			ui.DisplayNewline()
			Expect(out).To(Say("\n"))
		})
	})

	Describe("DisplayOK", func() {
		It("displays 'OK' in green and bold", func() {
			ui.DisplayOK()
			Expect(out).To(Say("\u001B\\[32;1mOK\u001B\\[0;22m"))
		})
	})

	// Covers the happy paths, additional cases are tested in TranslateText
	Describe("DisplayText", func() {
		It("displays the template with map values substituted into ui.Out with a newline", func() {
			ui.DisplayText(
				"template with {{.SomeMapValue}}",
				map[string]interface{}{
					"SomeMapValue": "map-value",
				})
			Expect(out).To(Say("template with map-value\n"))
		})

		When("the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("displays the translated template with map values substituted into ui.Out", func() {
				ui.DisplayText(
					"\nTIP: Use '{{.Command}}' to target new org",
					map[string]interface{}{
						"Command": "foo",
					})
				Expect(out).To(Say("\nASTUCE : utilisez 'foo' pour cibler une nouvelle organisation"))
			})
		})
	})

	Describe("DisplayTextLiteral", func() {
		It("displays the text into ui.Out with a newline", func() {
			ui.DisplayTextLiteral("some text")
			Expect(out).To(Say("some text\n"))
		})
	})

	Describe("Display JSON", func() {
		It("displays the indented JSON object", func() {
			obj := map[string]interface{}{
				"str":  "hello",
				"bool": true,
				"int":  42,
				"pass": "abc>&gd!f",
				"map":  map[string]interface{}{"float": 123.03},
				"arr":  []string{"a", "b"},
			}

			Expect(ui.DisplayJSON("named_json", obj)).To(Succeed())

			Expect(out).To(SatisfyAll(
				Say("named_json: {\n"),
				Say("  \"arr\": \\[\n"),
				Say("    \"a\","),
				Say("    \"b\"\n"),
				Say("  \\],\n"),
				Say("  \"bool\": true,\n"),
				Say("  \"int\": 42,\n"),
				Say("  \"map\": {\n"),
				Say("    \"float\": 123.03\n"),
				Say("  },\n"),
				Say("  \"pass\": \"abc>&gd!f\",\n"),
				Say("  \"str\": \"hello\"\n"),
				Say("}\n"),
				Say("\n"),
			))
		})
	})

	Describe("DeferText", func() {
		It("defers the template with map values substituted into ui.Out with a newline", func() {
			ui.DeferText(
				"template with {{.SomeMapValue}}",
				map[string]interface{}{
					"SomeMapValue": "map-value",
				})
			Expect(out).NotTo(Say("template with map-value\n"))
			ui.FlushDeferred()
			Expect(out).To(Say("template with map-value\n"))
		})

		When("the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("defers the translated template with map values substituted into ui.Out", func() {
				ui.DeferText(
					"\nTIP: Use '{{.Command}}' to target new org",
					map[string]interface{}{
						"Command": "foo",
					})
				Expect(out).NotTo(Say("\nASTUCE : utilisez 'foo' pour cibler une nouvelle organisation"))
				ui.FlushDeferred()
				Expect(out).To(Say("\nASTUCE : utilisez 'foo' pour cibler une nouvelle organisation"))
				ui.FlushDeferred()
				Expect(out).NotTo(Say("\nASTUCE : utilisez 'foo' pour cibler une nouvelle organisation"))
			})
		})
	})

	Describe("DisplayTextWithBold", func() {
		It("displays the template to ui.Out", func() {
			ui.DisplayTextWithBold("some-template")
			Expect(out).To(Say("some-template"))
		})

		When("an optional map is passed in", func() {
			It("displays the template with map values bolded and substituted into ui.Out", func() {
				ui.DisplayTextWithBold(
					"template with {{.SomeMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					})
				Expect(out).To(Say("template with \u001B\\[1mmap-value\u001B\\[22m"))
			})
		})

		When("multiple optional maps are passed in", func() {
			It("displays the template with only the first map values bolded and substituted into ui.Out", func() {
				ui.DisplayTextWithBold(
					"template with {{.SomeMapValue}} and {{.SomeOtherMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					},
					map[string]interface{}{
						"SomeOtherMapValue": "other-map-value",
					})
				Expect(out).To(Say("template with \u001B\\[1mmap-value\u001B\\[22m and <no value>"))
			})
		})

		When("the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("displays the translated template with map values bolded and substituted into ui.Out", func() {
				ui.DisplayTextWithBold(
					"App {{.AppName}} does not exist.",
					map[string]interface{}{
						"AppName": "some-app-name",
					})
				Expect(out).To(Say("L'application \u001B\\[1msome-app-name\u001B\\[22m n'existe pas.\n"))
			})
		})
	})

	Describe("DisplayTextWithFlavor", func() {
		It("displays the template to ui.Out", func() {
			ui.DisplayTextWithFlavor("some-template")
			Expect(out).To(Say("some-template"))
		})

		When("an optional map is passed in", func() {
			It("displays the template with map values colorized, bolded, and substituted into ui.Out", func() {
				ui.DisplayTextWithFlavor(
					"template with {{.SomeMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					})
				Expect(out).To(Say("template with \u001B\\[36;1mmap-value\u001B\\[0;22m"))
			})
		})

		When("multiple optional maps are passed in", func() {
			It("displays the template with only the first map values colorized, bolded, and substituted into ui.Out", func() {
				ui.DisplayTextWithFlavor(
					"template with {{.SomeMapValue}} and {{.SomeOtherMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					},
					map[string]interface{}{
						"SomeOtherMapValue": "other-map-value",
					})
				Expect(out).To(Say("template with \u001B\\[36;1mmap-value\u001B\\[0;22m and <no value>"))
			})
		})

		When("the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("displays the translated template with map values colorized, bolded and substituted into ui.Out", func() {
				ui.DisplayTextWithFlavor(
					"App {{.AppName}} does not exist.",
					map[string]interface{}{
						"AppName": "some-app-name",
					})
				Expect(out).To(Say("L'application \u001B\\[36;1msome-app-name\u001B\\[0;22m n'existe pas.\n"))
			})
		})
	})

	Describe("DisplayDiffAddition", func() {
		It("displays a green indented line with a +", func() {
			ui.DisplayDiffAddition("added", 3, false)
			Expect(out).To(Say("\u001B\\[32m\\+\\s+added\u001B\\[0m"))
		})
		It("displays a hyphen when the addHyphen is true", func() {
			ui.DisplayDiffAddition("added", 3, true)
			Expect(out).To(Say("\u001B\\[32m\\+\\s+\\- added\u001B\\[0m"))
		})

	})

	Describe("DisplayDiffRemoval", func() {
		It("displays a red indented line with a -", func() {
			ui.DisplayDiffRemoval("removed", 3, false)
			Expect(out).To(Say("\u001B\\[31m\\-\\s+removed\u001B\\[0m"))
		})
		It("displays a a hyphen when addHyphen is true", func() {
			ui.DisplayDiffRemoval("removed", 3, true)
			Expect(out).To(Say("\u001B\\[31m\\-\\s+\\- removed\u001B\\[0m"))
		})
	})

	Describe("DisplayDiffUnchanged", func() {
		It("displays a plain indented line with no prefix", func() {
			ui.DisplayDiffUnchanged("unchanged", 3, false)
			Expect(out).To(Say("        unchanged"))
		})
		It("displays a a hyphen when addHyphen is true", func() {
			ui.DisplayDiffUnchanged("unchanged", 3, true)
			Expect(out).To(Say("      - unchanged"))
		})
	})

	Describe("TranslateText", func() {
		It("returns the template", func() {
			Expect(ui.TranslateText("some-template")).To(Equal("some-template"))
		})

		When("an optional map is passed in", func() {
			It("returns the template with map values substituted in", func() {
				expected := ui.TranslateText(
					"template {{.SomeMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					})
				Expect(expected).To(Equal("template map-value"))
			})
		})

		When("multiple optional maps are passed in", func() {
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

		When("the locale is not set to english", func() {
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

	Describe("UserFriendlyDate", func() {
		It("formats a time into an ISO8601 string", func() {
			Expect(ui.UserFriendlyDate(time.Unix(0, 0))).To(MatchRegexp(`\w{3} [0-3]\d \w{3} [0-2]\d:[0-5]\d:[0-5]\d \w+ \d{4}`))
		})
	})
})
