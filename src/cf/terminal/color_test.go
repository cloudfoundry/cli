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

var _ = Describe("Terminal colours", func() {
	Describe("CF_COLOR", func() {
		Context("On OSes that don't support colours", func() {
			BeforeEach(func() { OsSupportsColours = false })

			Context("When the CF_COLOR env variable is specified", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "true") })
				itDoesntColourize()
			})

			Context("When the CF_COLOR env variable is not specified", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "") })
				itDoesntColourize()
			})
		})

		Context("On OSes that support colours", func() {
			BeforeEach(func() { OsSupportsColours = true })

			Context("When the CF_COLOR env variable is not specified", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "") })

				Context("And the terminal supports colours", func() {
					BeforeEach(func() { TerminalSupportsColours = true })
					itColourizes()
				})

				Context("And the terminal doesn't support colours", func() {
					BeforeEach(func() { TerminalSupportsColours = false })
					itDoesntColourize()
				})
			})

			Context("When the CF_COLOR env variable is set to 'true'", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "true") })

				Context("And the terminal supports colours", func() {
					BeforeEach(func() { TerminalSupportsColours = true })
					itColourizes()
				})

				Context("Even if the terminal doesn't support colours", func() {
					BeforeEach(func() { TerminalSupportsColours = false })
					itColourizes()
				})
			})

			Context("When the CF_COLOR env variable is set to 'false', even if the terminal supports colours", func() {
				BeforeEach(func() {
					os.Setenv("CF_COLOR", "false")
					TerminalSupportsColours = true
				})

				itDoesntColourize()
			})
		})
	})

	Describe("OsSupportsColours", func() {
		It("Returns false on windows, and true otherwise", func() {
			if runtime.GOOS == "windows" {
				Expect(OsSupportsColours).To(BeFalse())
			} else {
				Expect(OsSupportsColours).To(BeTrue())
			}
		})
	})

	var (
		originalOsSupportsColours       bool
		originalTerminalSupportsColours bool
	)

	BeforeEach(func() {
		originalOsSupportsColours = OsSupportsColours
		originalTerminalSupportsColours = TerminalSupportsColours
	})

	AfterEach(func() {
		OsSupportsColours = originalOsSupportsColours
		TerminalSupportsColours = originalTerminalSupportsColours
		os.Setenv("CF_COLOR", "false")
	})
})

func itColourizes() {
	It("colourizes", func() {
		text := "Hello World"
		colorizedText := Colorize(text, 31, true)
		Expect(colorizedText).To(Equal("\033[1;31mHello World\033[0m"))
	})
}

func itDoesntColourize() {
	It("doesn't colourize", func() {
		text := "Hello World"
		colorizedText := Colorize(text, 31, true)
		Expect(colorizedText).To(Equal("Hello World"))
	})
}
