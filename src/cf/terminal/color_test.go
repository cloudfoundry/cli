/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package terminal_test

import (
	. "cf/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"runtime"
)

var _ = Describe("Terminal colors", func() {
	JustBeforeEach(func() {
		ResetColorSupport()
	})

	Describe("CF_COLOR", func() {
		Context("On OSes that don't support colors", func() {
			BeforeEach(func() { OsSupportsColors = false })

			Context("When the CF_COLOR env variable is specified", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "true") })
				itDoesntColorize()
			})

			Context("When the CF_COLOR env variable is not specified", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "") })
				itDoesntColorize()
			})
		})

		Context("On OSes that support colors", func() {
			BeforeEach(func() { OsSupportsColors = true })

			Context("When the CF_COLOR env variable is not specified", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "") })

				Context("And the terminal supports colors", func() {
					BeforeEach(func() { TerminalSupportsColors = true })
					itColorizes()
				})

				Context("And the terminal doesn't support colors", func() {
					BeforeEach(func() { TerminalSupportsColors = false })
					itDoesntColorize()
				})
			})

			Context("When the CF_COLOR env variable is set to 'true'", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "true") })

				Context("And the terminal supports colors", func() {
					BeforeEach(func() { TerminalSupportsColors = true })
					itColorizes()
				})

				Context("Even if the terminal doesn't support colors", func() {
					BeforeEach(func() { TerminalSupportsColors = false })
					itColorizes()
				})
			})

			Context("When the CF_COLOR env variable is set to 'false', even if the terminal supports colors", func() {
				BeforeEach(func() {
					os.Setenv("CF_COLOR", "false")
					TerminalSupportsColors = true
				})

				itDoesntColorize()
			})
		})
	})

	Describe("OsSupportsColors", func() {
		It("Returns false on windows, and true otherwise", func() {
			if runtime.GOOS == "windows" {
				Expect(OsSupportsColors).To(BeFalse())
			} else {
				Expect(OsSupportsColors).To(BeTrue())
			}
		})
	})

	var (
		originalOsSupportsColors       bool
		originalTerminalSupportsColors bool
	)

	BeforeEach(func() {
		originalOsSupportsColors = OsSupportsColors
		originalTerminalSupportsColors = TerminalSupportsColors
	})

	AfterEach(func() {
		OsSupportsColors = originalOsSupportsColors
		TerminalSupportsColors = originalTerminalSupportsColors
		os.Setenv("CF_COLOR", "false")
	})
})

func itColorizes() {
	It("colorizes", func() {
		text := "Hello World"
		colorizedText := ColorizeBold(text, 31)
		Expect(colorizedText).To(Equal("\033[1;31mHello World\033[0m"))
	})
}

func itDoesntColorize() {
	It("doesn't colorize", func() {
		text := "Hello World"
		colorizedText := ColorizeBold(text, 31)
		Expect(colorizedText).To(Equal("Hello World"))
	})
}
