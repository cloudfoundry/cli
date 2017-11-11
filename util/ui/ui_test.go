package ui_test

import (
	"errors"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/command/translatableerror/translatableerrorfakes"
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
		ui.Err = NewBuffer()
	})

	It("sets the TimezoneLocation to the local timezone", func() {
		location := time.Now().Location()
		Expect(ui.TimezoneLocation).To(Equal(location))
	})

	Describe("DisplayPasswordPrompt", func() {
		var inBuffer *Buffer

		BeforeEach(func() {
			inBuffer = NewBuffer()
			ui.In = inBuffer
			inBuffer.Write([]byte("some-input\n"))
		})

		It("displays the passed in string", func() {
			_, _ = ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
				"AppName": "some-app",
			})
			Expect(out).To(Say("App some-app does not exist."))
		})

		It("returns the user input", func() {
			userInput, err := ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
				"AppName": "some-app",
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(userInput).To(Equal("some-input"))
			Expect(out).ToNot(Say("some-input"))
		})

		Context("when the locale is not set to English", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
				ui.OutForInteration = out
			})

			It("translates and displays the prompt", func() {
				_, _ = ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
					"AppName": "some-app",
				})
				Expect(out).To(Say("L'application some-app n'existe pas.\n"))
			})
		})
	})

	Describe("DisplayBoolPrompt", func() {
		var inBuffer *Buffer

		BeforeEach(func() {
			inBuffer = NewBuffer()
			ui.In = inBuffer
		})

		It("displays the passed in string", func() {
			_, _ = ui.DisplayBoolPrompt(false, "some-prompt", nil)
			Expect(out).To(Say("some-prompt \\[yN\\]:"))
		})

		Context("when the user chooses yes", func() {
			BeforeEach(func() {
				_, err := inBuffer.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns true", func() {
				response, err := ui.DisplayBoolPrompt(false, "some-prompt", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(BeTrue())
			})
		})

		Context("when the user chooses no", func() {
			BeforeEach(func() {
				_, err := inBuffer.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns false", func() {
				response, err := ui.DisplayBoolPrompt(false, "some-prompt", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(BeFalse())
			})
		})

		Context("when the user chooses the default", func() {
			BeforeEach(func() {
				_, err := inBuffer.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when the default is true", func() {
				It("returns true", func() {
					response, err := ui.DisplayBoolPrompt(true, "some-prompt", nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeTrue())
				})
			})

			Context("when the default is false", func() {
				It("returns false", func() {
					response, err := ui.DisplayBoolPrompt(false, "some-prompt", nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeFalse())
				})
			})
		})

		Context("when the interact library returns an error", func() {
			It("returns the error", func() {
				_, err := inBuffer.Write([]byte("invalid\n"))
				Expect(err).ToNot(HaveOccurred())
				_, err = ui.DisplayBoolPrompt(false, "some-prompt", nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DisplayError", func() {
		Context("when passed a TranslatableError", func() {
			var fakeTranslateErr *translatableerrorfakes.FakeTranslatableError

			BeforeEach(func() {
				fakeTranslateErr = new(translatableerrorfakes.FakeTranslatableError)
				fakeTranslateErr.TranslateReturns("I am an error")

				ui.DisplayError(fakeTranslateErr)
			})

			It("displays the error to ui.Err and displays FAILED in bold red to ui.Out", func() {
				Expect(ui.Err).To(Say("I am an error\n"))
				Expect(out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
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
				Expect(out).To(Say("\x1b\\[31;1mFAILED\x1b\\[0m\n"))
			})
		})
	})

	Describe("DisplayHeader", func() {
		It("displays the header colorized and bolded to ui.Out", func() {
			ui.DisplayHeader("some-header")
			Expect(out).To(Say("\x1b\\[1msome-header\x1b\\[0m"))
		})

		Context("when the locale is not set to English", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("displays the translated header colorized and bolded to ui.Out", func() {
				ui.DisplayHeader("FEATURE FLAGS")
				Expect(out).To(Say("\x1b\\[1mINDICATEURS DE FONCTION\x1b\\[0m"))
			})
		})
	})

	Describe("DisplayKeyValueTable", func() {
		JustBeforeEach(func() {
			ui.DisplayKeyValueTable(" ",
				[][]string{
					{"wut0:", ""},
					{"wut1:", "hi hi"},
					{"wut2:", strings.Repeat("a", 9)},
					{"wut3:", "hi hi " + strings.Repeat("a", 9)},
					{"wut4:", strings.Repeat("a", 15) + " " + strings.Repeat("b", 15)},
				},
				2)
		})

		Context("in a TTY", func() {
			BeforeEach(func() {
				ui.IsTTY = true
				ui.TerminalWidth = 20
			})

			It("displays a table with the last column wrapping according to width", func() {
				Expect(out).To(Say(" wut0:  " + "\n"))
				Expect(out).To(Say(" wut1:  " + "hi hi\n"))
				Expect(out).To(Say(" wut2:  " + strings.Repeat("a", 9) + "\n"))
				Expect(out).To(Say(" wut3:  hi hi\n"))
				Expect(out).To(Say("        " + strings.Repeat("a", 9) + "\n"))
				Expect(out).To(Say(" wut4:  " + strings.Repeat("a", 15) + "\n"))
				Expect(out).To(Say("        " + strings.Repeat("b", 15) + "\n"))
			})
		})
	})

	Describe("DisplayLogMessage", func() {
		var message *uifakes.FakeLogMessage

		BeforeEach(func() {
			var err error
			ui.TimezoneLocation, err = time.LoadLocation("America/Los_Angeles")
			Expect(err).NotTo(HaveOccurred())

			message = new(uifakes.FakeLogMessage)
			message.MessageReturns("This is a log message\r\n")
			message.TypeReturns("OUT")
			message.TimestampReturns(time.Unix(1468969692, 0)) // "2016-07-19T16:08:12-07:00"
			message.SourceTypeReturns("APP/PROC/WEB")
			message.SourceInstanceReturns("12")
		})

		Context("with header", func() {
			Context("single line log message", func() {
				It("prints out a single line to STDOUT", func() {
					ui.DisplayLogMessage(message, true)
					Expect(out).To(Say("2016-07-19T16:08:12.00-0700 \\[APP/PROC/WEB/12\\] OUT This is a log message\n"))
				})
			})

			Context("multi-line log message", func() {
				BeforeEach(func() {
					var err error
					ui.TimezoneLocation, err = time.LoadLocation("America/Los_Angeles")
					Expect(err).NotTo(HaveOccurred())

					message.MessageReturns("This is a log message\nThis is also a log message")
				})

				It("prints out mutliple lines to STDOUT", func() {
					ui.DisplayLogMessage(message, true)
					Expect(out).To(Say("2016-07-19T16:08:12.00-0700 \\[APP/PROC/WEB/12\\] OUT This is a log message\n"))
					Expect(out).To(Say("2016-07-19T16:08:12.00-0700 \\[APP/PROC/WEB/12\\] OUT This is also a log message\n"))
				})
			})
		})

		Context("without header", func() {
			Context("single line log message", func() {
				It("prints out a single line to STDOUT", func() {
					ui.DisplayLogMessage(message, false)
					Expect(out).To(Say("This is a log message\n"))
				})
			})

			Context("multi-line log message", func() {
				BeforeEach(func() {
					var err error
					ui.TimezoneLocation, err = time.LoadLocation("America/Los_Angeles")
					Expect(err).NotTo(HaveOccurred())

					message.MessageReturns("This is a log message\nThis is also a log message")
				})

				It("prints out mutliple lines to STDOUT", func() {
					ui.DisplayLogMessage(message, false)
					Expect(out).To(Say("This is a log message\n"))
					Expect(out).To(Say("This is also a log message\n"))
				})
			})
		})

		Context("error log lines", func() {
			BeforeEach(func() {
				message.TypeReturns("ERR")
			})
			It("colors the line red", func() {
				ui.DisplayLogMessage(message, false)
				Expect(out).To(Say("\x1b\\[31mThis is a log message\x1b\\[0m\n"))
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
			Expect(out).To(Say("\x1b\\[32;1mOK\x1b\\[0m"))
		})
	})

	Describe("DisplayTableWithHeader", func() {
		It("makes the first row bold", func() {
			ui.DisplayTableWithHeader(" ",
				[][]string{
					{"", "header1", "header2", "header3"},
					{"#0", "data1", "data2", "data3"},
				},
				2)
			Expect(out).To(Say("    \x1b\\[1mheader1\x1b\\[0m")) // Makes sure empty values are not bolded
			Expect(out).To(Say("\x1b\\[1mheader2\x1b\\[0m"))
			Expect(out).To(Say("\x1b\\[1mheader3\x1b\\[0m"))
			Expect(out).To(Say("#0  data1    data2    data3"))
		})
	})

	// Covers the happy paths, additional cases are tested in TranslateText
	Describe("DisplayText", func() {
		It("displays the template with map values substituted in to ui.Out with a newline", func() {
			ui.DisplayText(
				"template with {{.SomeMapValue}}",
				map[string]interface{}{
					"SomeMapValue": "map-value",
				})
			Expect(out).To(Say("template with map-value\n"))
		})

		Context("when the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("displays the translated template with map values substituted in to ui.Out", func() {
				ui.DisplayText(
					"\nTIP: Use '{{.Command}}' to target new org",
					map[string]interface{}{
						"Command": "foo",
					})
				Expect(out).To(Say("\nASTUCE : utilisez 'foo' pour cibler une nouvelle organisation"))
			})
		})
	})

	Describe("DisplayTextWithFlavor", func() {
		It("displays the template to ui.Out", func() {
			ui.DisplayTextWithFlavor("some-template")
			Expect(out).To(Say("some-template"))
		})

		Context("when an optional map is passed in", func() {
			It("displays the template with map values colorized, bolded, and substituted in to ui.Out", func() {
				ui.DisplayTextWithFlavor(
					"template with {{.SomeMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					})
				Expect(out).To(Say("template with \x1b\\[36;1mmap-value\x1b\\[0m"))
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
				Expect(out).To(Say("template with \x1b\\[36;1mmap-value\x1b\\[0m and <no value>"))
			})
		})

		Context("when the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("displays the translated template with map values colorized, bolded and substituted in to ui.Out", func() {
				ui.DisplayTextWithFlavor(
					"App {{.AppName}} does not exist.",
					map[string]interface{}{
						"AppName": "some-app-name",
					})
				Expect(out).To(Say("L'application \x1b\\[36;1msome-app-name\x1b\\[0m n'existe pas.\n"))
			})
		})
	})

	Describe("DisplayTextWithBold", func() {
		It("displays the template to ui.Out", func() {
			ui.DisplayTextWithBold("some-template")
			Expect(out).To(Say("some-template"))
		})

		Context("when an optional map is passed in", func() {
			It("displays the template with map values bolded and substituted in to ui.Out", func() {
				ui.DisplayTextWithBold(
					"template with {{.SomeMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					})
				Expect(out).To(Say("template with \x1b\\[1mmap-value\x1b\\[0m"))
			})
		})

		Context("when multiple optional maps are passed in", func() {
			It("displays the template with only the first map values bolded and substituted in to ui.Out", func() {
				ui.DisplayTextWithBold(
					"template with {{.SomeMapValue}} and {{.SomeOtherMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "map-value",
					},
					map[string]interface{}{
						"SomeOtherMapValue": "other-map-value",
					})
				Expect(out).To(Say("template with \x1b\\[1mmap-value\x1b\\[0m and <no value>"))
			})
		})

		Context("when the locale is not set to english", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
			})

			It("displays the translated template with map values bolded and substituted in to ui.Out", func() {
				ui.DisplayTextWithBold(
					"App {{.AppName}} does not exist.",
					map[string]interface{}{
						"AppName": "some-app-name",
					})
				Expect(out).To(Say("L'application \x1b\\[1msome-app-name\x1b\\[0m n'existe pas.\n"))
			})
		})
	})

	// Covers the happy paths, additional cases are tested in TranslateText
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

	// Covers the happy paths, additional cases are tested in TranslateText
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

	Describe("RequestLoggerFileWriter", func() {
		It("returns a RequestLoggerFileWriter with the consistent filewriting mutex", func() {
			logger1 := ui.RequestLoggerFileWriter(nil)
			logger2 := ui.RequestLoggerFileWriter(nil)

			c := make(chan bool)
			err := logger1.Start()
			Expect(err).ToNot(HaveOccurred())
			go func() {
				Expect(logger2.Start()).ToNot(HaveOccurred())
				c <- true
			}()
			Consistently(c).ShouldNot(Receive())
			Expect(logger1.Stop()).ToNot(HaveOccurred())
			Eventually(c).Should(Receive())
		})
	})

	Describe("RequestLoggerTerminalDisplay", func() {
		It("returns a RequestLoggerTerminalDisplay with the consistent display mutex", func() {
			logger1 := ui.RequestLoggerTerminalDisplay()
			logger2 := ui.RequestLoggerTerminalDisplay()

			c := make(chan bool)
			err := logger1.Start()
			Expect(err).ToNot(HaveOccurred())
			go func() {
				Expect(logger2.Start()).ToNot(HaveOccurred())
				c <- true
			}()
			Consistently(c).ShouldNot(Receive())
			Expect(logger1.Stop()).ToNot(HaveOccurred())
			Eventually(c).Should(Receive())
		})
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

	Describe("UserFriendlyDate", func() {
		It("formats a time into an ISO8601 string", func() {
			Expect(ui.UserFriendlyDate(time.Unix(0, 0))).To(MatchRegexp("\\w{3} [0-3]\\d \\w{3} [0-2]\\d:[0-5]\\d:[0-5]\\d \\w+ \\d{4}"))
		})
	})
})
