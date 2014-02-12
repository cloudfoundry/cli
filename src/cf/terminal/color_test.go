package terminal_test

import (
	. "cf/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"runtime"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestColorize", func() {
		os.Setenv("CF_COLOR", "true")
		text := "Hello World"
		colorizedText := Colorize(text, 31, true)

		if runtime.GOOS == "windows" {
			Expect(colorizedText).To(Equal("Hello World"))
		} else {
			Expect(colorizedText).To(Equal("\033[1;31mHello World\033[0m"))
		}
	})
})
