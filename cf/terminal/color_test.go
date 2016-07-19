package terminal_test

import (
	"os"

	. "code.cloudfoundry.org/cli/cf/terminal"
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Terminal colors", func() {
	BeforeEach(func() {
		UserAskedForColors = ""
	})

	JustBeforeEach(func() {
		InitColorSupport()
	})

	Describe("CF_COLOR", func() {
		Context("All OSes support colors", func() {
			Context("When the CF_COLOR env variable is not specified", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "") })

				Context("And the terminal supports colors", func() {
					BeforeEach(func() { TerminalSupportsColors = true })
					itColorizes()

					Context("And user does not ask for color", func() {
						BeforeEach(func() { UserAskedForColors = "false" })
						itDoesntColorize()
					})

					Context("And user does ask for color", func() {
						BeforeEach(func() { UserAskedForColors = "true" })
						itColorizes()
					})
				})

				Context("And the terminal doesn't support colors", func() {
					BeforeEach(func() { TerminalSupportsColors = false })
					itDoesntColorize()

					Context("And user asked for color", func() {
						BeforeEach(func() { UserAskedForColors = "true" })
						itColorizes()
					})
				})
			})

			Context("When the CF_COLOR env variable is set to 'true'", func() {
				BeforeEach(func() { os.Setenv("CF_COLOR", "true") })

				Context("And the terminal supports colors", func() {
					BeforeEach(func() { TerminalSupportsColors = true })
					itColorizes()

					Context("and the user asked for colors", func() {
						BeforeEach(func() { UserAskedForColors = "true" })
						itColorizes()
					})

					Context("and the user did not ask for colors", func() {
						BeforeEach(func() { UserAskedForColors = "false" })
						itColorizes()
					})
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

				Context("and the user asked for colors", func() {
					BeforeEach(func() { UserAskedForColors = "true" })
					itDoesntColorize()
				})
			})
		})
	})

	var (
		originalTerminalSupportsColors bool
	)

	BeforeEach(func() {
		originalTerminalSupportsColors = TerminalSupportsColors
	})

	AfterEach(func() {
		TerminalSupportsColors = originalTerminalSupportsColors
		os.Setenv("CF_COLOR", "false")
	})
})

func itColorizes() {
	It("colorizes", func() {
		text := "Hello World"
		colorizedText := ColorizeBold(text, 33)
		colorizeYellow := color.New(color.FgYellow).Add(color.Bold).SprintFunc()
		Expect(colorizedText).To(Equal(colorizeYellow(text)))
	})
}

func itDoesntColorize() {
	It("doesn't colorize", func() {
		text := "Hello World"
		colorizedText := ColorizeBold(text, 33)
		Expect(colorizedText).To(Equal("Hello World"))
	})
}
